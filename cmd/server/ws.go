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
 return &wsAPI{auth:a,connections:connections,slots:make(chan struct{},wsMaxConcurrent),roles:roles,approvals:approvals,delivery:delivery,router:router}
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
  for {
   kind,_,err:=conn.ReadMessage()
   if err!=nil{return}
   // Reject all application messages. The socket is identity-only.
   if kind==websocket.TextMessage||kind==websocket.BinaryMessage{
    _=conn.WriteControl(websocket.CloseMessage,websocket.FormatCloseMessage(websocket.ClosePolicyViolation,"signaling not enabled"),time.Now().Add(5*time.Second))
    return
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
