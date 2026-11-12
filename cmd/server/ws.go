package main

import (
 "crypto/ed25519"
 "encoding/base64"
 "encoding/json"
 "net/http"
 "sync"
 "time"

 "github.com/gorilla/websocket"
 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/Ploos-AS/PloosAssist/internal/signaling"
 "github.com/Ploos-AS/PloosAssist/internal/session"
)

const (
 wsHandshakeTimeout = 15*time.Second
 wsIdleTimeout = 60*time.Second
 wsPingInterval = 20*time.Second
 wsMaxMessage = 4096
 wsMaxConcurrent = 128
)

// wsAPI is a loopback-only authenticated transport experiment.
// Authentication does NOT imply session membership or host consent.
// No signaling messages are accepted or forwarded.
type wsAPI struct {
 auth *identity.TransportAuth
 connections *signaling.Connections
 slots chan struct{}
 roles *signaling.Roles
 approvals *signaling.ApprovalRequests
 delivery *signaling.HostDelivery
 router *signaling.Router
 enrollment *signaling.Enrollment
 revocation *signaling.Revocation
 activeMu sync.Mutex
 active map[string]activeSocket
}
type activeSocket struct{session string;conn *websocket.Conn}
func (a *wsAPI)track(id,session string,conn *websocket.Conn){a.activeMu.Lock();defer a.activeMu.Unlock();a.active[id]=activeSocket{session,conn}}
func (a *wsAPI)untrack(id string){a.activeMu.Lock();defer a.activeMu.Unlock();delete(a.active,id)}
func (a *wsAPI)terminateSession(session string){
 a.activeMu.Lock()
 var sockets []*websocket.Conn
 for id,s:=range a.active{if s.session==session{sockets=append(sockets,s.conn);delete(a.active,id)}}
 a.activeMu.Unlock()
 for _,conn:=range sockets{_ =conn.Close()}
}
func newWSAPI()*wsAPI {
 a,err:=identity.NewTransportAuth(30*time.Second,wsMaxConcurrent)
 if err!=nil{panic(err)}
 connections,err:=signaling.NewConnections(wsMaxConcurrent,2*time.Minute)
 if err!=nil{panic(err)}
 consent,_:=signaling.NewConsentPolicy(wsMaxConcurrent)
 roles,_:=signaling.NewRoles(wsMaxConcurrent,connections)
 coordinator,_:=signaling.NewCoordinator(connections,consent)
 approvals,_:=signaling.NewApprovalRequests(wsMaxConcurrent,connections,coordinator)
 events,_:=signaling.NewApprovalEvents(wsMaxConcurrent,approvals)
 delivery,_:=signaling.NewHostDelivery(wsMaxConcurrent,connections,approvals,events)
 router,_:=signaling.NewRouter(connections,consent)
 return &wsAPI{auth:a,connections:connections,slots:make(chan struct{},wsMaxConcurrent),active:make(map[string]activeSocket),roles:roles,approvals:approvals,delivery:delivery,router:router,revocation:nil}
}

func mustEnrollment(store *session.Store,connections *signaling.Connections,roles *signaling.Roles)*signaling.Enrollment{
 e,err:=signaling.NewEnrollment(store,connections,roles)
 if err!=nil{panic(err)}
 return e
}

func (a *wsAPI)routes(mux *http.ServeMux) {
 mux.HandleFunc("GET /v1/ws",func(w http.ResponseWriter,r *http.Request){
  key,err:=base64.RawURLEncoding.DecodeString(r.URL.Query().Get("public_key"))
  sessionID:=r.URL.Query().Get("session_id")
  if err!=nil||len(key)!=ed25519.PublicKeySize||len(sessionID)==0||len(sessionID)>128{
   jsonReply(w,http.StatusBadRequest,map[string]string{"error":"invalid identity parameters"});return
  }
  select {
  case a.slots<-struct{}{}:
   defer func(){<-a.slots}()
  default:
   jsonReply(w,http.StatusServiceUnavailable,map[string]string{"error":"connection capacity reached"});return
  }
  upgrader:=websocket.Upgrader{
   HandshakeTimeout:5*time.Second,
   CheckOrigin:func(r *http.Request)bool{return r.Header.Get("Origin")==""},
  }
  conn,err:=upgrader.Upgrade(w,r,nil)
  if err!=nil{return}
  defer conn.Close()
  conn.SetReadLimit(wsMaxMessage)
  _=conn.SetReadDeadline(time.Now().Add(wsHandshakeTimeout))
  id,err:=a.auth.Open(sessionID,ed25519.PublicKey(key))
  if err!=nil{return}
  defer a.auth.Close(id)
  defer a.connections.Close(id)
  // Cleanup must run while the verified connection identity still exists.
  // Never trust a disconnect message supplied by the peer.
  defer a.cleanupConnection(id)
  defer a.untrack(id)
  challengeID,ch,err:=a.auth.Challenge(id)
  if err!=nil{return}
  start,_:=json.Marshal(map[string]string{"type":"challenge","nonce":base64.RawURLEncoding.EncodeToString(ch.Nonce),"challenge_id":challengeID})
  if err:=conn.WriteMessage(websocket.TextMessage,start);err!=nil{return}
  kind,payload,err:=conn.ReadMessage()
  if err!=nil||kind!=websocket.TextMessage{return}
  var proof struct{Type string `json:"type"`;Signature string `json:"signature"`}
  if json.Unmarshal(payload,&proof)!=nil||proof.Type!="proof"{return}
  sig,err:=base64.RawURLEncoding.DecodeString(proof.Signature)
  if err!=nil{return}
  fingerprint,err:=a.auth.Authenticate(id,challengeID,sig)
  if err!=nil{return}
  verifiedSession,verifiedFingerprint,err:=a.auth.Identity(id)
  if err!=nil||verifiedSession!=sessionID||verifiedFingerprint!=fingerprint{return}
  if err:=a.connections.Register(id,verifiedSession,verifiedFingerprint);err!=nil{return}
  a.track(id,verifiedSession,conn)
  response,_:=json.Marshal(map[string]string{"type":"proof_verified","fingerprint":fingerprint})
  if err:=conn.WriteMessage(websocket.TextMessage,response);err!=nil{return}
  // Keep the authenticated socket alive, but never accept application
  // messages until pairing and consent are bound to this exact connection.
  _=conn.SetReadDeadline(time.Now().Add(wsIdleTimeout))
  conn.SetPongHandler(func(string)error{
   if err:=a.connections.Refresh(id);err!=nil{return err}
   return conn.SetReadDeadline(time.Now().Add(wsIdleTimeout))
  })
  stop:=make(chan struct{})
  var once sync.Once
  defer once.Do(func(){close(stop)})
  go func(){
   ticker:=time.NewTicker(wsPingInterval)
   defer ticker.Stop()
   for {
    select {
    case <-stop:return
    case <-ticker.C:
     if err:=conn.WriteControl(websocket.PingMessage,nil,time.Now().Add(5*time.Second));err!=nil{
      _=conn.Close()
      return
     }
    }
   }
  }()
  enrolled:=false
  var enrolledRole signaling.Role
  for {
   kind,payload,err:=conn.ReadMessage()
   if err!=nil{return}
   if kind!=websocket.TextMessage{a.wsPolicyClose(conn);return}
   if !enrolled{
    var request struct{
     Type string `json:"type"`
     Role signaling.Role `json:"role"`
     Capability string `json:"capability"`
    }
    if a.enrollment==nil||json.Unmarshal(payload,&request)!=nil||request.Type!="role_enroll"||len(request.Capability)>256||
     (request.Role!=signaling.HostRole&&request.Role!=signaling.TechnicianRole)||
     a.enrollment.Bind(sessionID,id,request.Role,request.Capability)!=nil{
     a.wsPolicyClose(conn);return
    }
    enrolled=true
    enrolledRole=request.Role
    ack,_:=json.Marshal(map[string]string{"type":"role_enrolled","role":string(request.Role)})
    if err:=conn.WriteMessage(websocket.TextMessage,ack);err!=nil{return}
    continue
   }
   var action struct{Type string `json:"type"`;TicketID string `json:"ticket_id"`;Approve *bool `json:"approve"`;Capability string `json:"capability"`}
   if json.Unmarshal(payload,&action)!=nil{a.wsPolicyClose(conn);return}
   switch {
   case action.Type=="approval_request"&&enrolledRole==signaling.TechnicianRole:
    hostID,techID,err:=a.roles.Pair(sessionID)
    if err!=nil||techID!=id{a.wsPolicyClose(conn);return}
    req,err:=a.approvals.Request(sessionID,hostID,id,2*time.Minute)
    if err!=nil{a.wsPolicyClose(conn);return}
    ticket,err:=a.delivery.Queue(req,hostID)
    if err!=nil{
     a.approvals.Cancel(sessionID)
     a.wsPolicyClose(conn);return
    }
    ack,_:=json.Marshal(map[string]string{"type":"approval_requested","expires_at":ticket.ExpiresAt.UTC().Format(time.RFC3339)})
    if err:=conn.WriteMessage(websocket.TextMessage,ack);err!=nil{return}
   case action.Type=="approval_check"&&enrolledRole==signaling.HostRole:
    hostID,_,err:=a.roles.Pair(sessionID)
    if err!=nil||hostID!=id{a.wsPolicyClose(conn);return}
    ticket,err:=a.delivery.Take(id)
    if err!=nil{
     ack,_:=json.Marshal(map[string]string{"type":"approval_none"})
     if err:=conn.WriteMessage(websocket.TextMessage,ack);err!=nil{return}
     continue
    }
    ack,_:=json.Marshal(map[string]string{
     "type":"approval_pending",
     "ticket_id":ticket.ID,
     "technician_fingerprint":ticket.TechnicianFingerprint,
     "expires_at":ticket.ExpiresAt.UTC().Format(time.RFC3339),
    })
    if err:=conn.WriteMessage(websocket.TextMessage,ack);err!=nil{return}
   case action.Type=="approval_decide"&&enrolledRole==signaling.HostRole:
    hostID,_,err:=a.roles.Pair(sessionID)
    if err!=nil||hostID!=id||action.Approve==nil||action.TicketID==""||len(action.TicketID)>128{
     a.wsPolicyClose(conn);return
    }
    // The ticket is one-use and bound to the verified host connection.
    // A decision is an explicit host action; retrieval never implies consent.
    if err:=a.delivery.Decide(id,action.TicketID,*action.Approve);err!=nil{
     a.wsPolicyClose(conn);return
    }
    state:="denied"
    if *action.Approve{state="approved"}
    ack,_:=json.Marshal(map[string]string{"type":"approval_decided","state":state})
    if err:=conn.WriteMessage(websocket.TextMessage,ack);err!=nil{return}
   case action.Type=="session_revoke"&&enrolledRole==signaling.HostRole:
    if a.revocation==nil||len(action.Capability)>256||action.Capability==""{
     a.wsPolicyClose(conn);return
    }
    // Revocation requires the independent host capability, not merely a
    // claimed role. Terminate both sockets after all in-memory grants close.
    if err:=a.revocation.Revoke(sessionID,action.Capability);err!=nil{a.wsPolicyClose(conn);return}
    a.terminateSession(sessionID)
    return
   default:
    a.wsPolicyClose(conn);return
   }
  }
 })
}

// cleanupConnection runs on every authenticated or partially authenticated
// socket exit, before its identity is removed from the connection registry.
func (a *wsAPI)cleanupConnection(connectionID string){
 a.router.Detach(connectionID)
 a.delivery.Disconnect(connectionID)
 a.roles.Disconnect(connectionID)
}

func (a *wsAPI)wsPolicyClose(conn *websocket.Conn){
 _=conn.WriteControl(websocket.CloseMessage,websocket.FormatCloseMessage(websocket.ClosePolicyViolation,"message not authorized"),time.Now().Add(5*time.Second))
}
