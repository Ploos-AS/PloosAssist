package signaling

import (
 "errors"
 "sync"
)

var ErrPeerOffline = errors.New("signaling recipient is offline")

// Router authorizes every message using an authenticated connection identity
// and a currently approved host/technician pair. This is an internal component:
// transport and host-consent verification MUST be provided by its caller.
type Router struct {
 mu sync.RWMutex
 connections *Connections
 consent *ConsentPolicy
 destinations map[string]map[string]chan Envelope
}

func NewRouter(connections *Connections, consent *ConsentPolicy)(*Router,error){
 if connections==nil||consent==nil{return nil,ErrConsentInvalid}
 return &Router{connections:connections,consent:consent,destinations:make(map[string]map[string]chan Envelope)},nil
}

// Attach associates a previously authenticated, live connection with a
// bounded outbound queue. No client-supplied fingerprint is accepted.
func (r *Router)Attach(connectionID string)(<-chan Envelope,error){
 session,fingerprint,err:=r.connections.Identity(connectionID)
 if err!=nil{return nil,err}
 r.mu.Lock()
 defer r.mu.Unlock()
 if r.destinations[session]==nil{r.destinations[session]=make(map[string]chan Envelope)}
 if _,exists:=r.destinations[session][fingerprint];exists{return nil,ErrConnectionLimit}
 queue:=make(chan Envelope,16)
 r.destinations[session][fingerprint]=queue
 return queue,nil
}

// Detach is called on disconnect. The caller must use the original
// authenticated connection ID, never a user-provided fingerprint.
func (r *Router)Detach(connectionID string){
 session,fingerprint,err:=r.connections.Identity(connectionID)
 if err!=nil{return}
 r.mu.Lock()
 defer r.mu.Unlock()
 if peers:=r.destinations[session];peers!=nil{
  delete(peers,fingerprint)
  if len(peers)==0{delete(r.destinations,session)}
 }
}

// Route uses the authenticated connection identity, not the sender field,
// and checks consent on every message, including after revocation.
func (r *Router)Route(connectionID string,e Envelope)error{
 session,fingerprint,err:=r.connections.Identity(connectionID)
 if err!=nil{return err}
 if err:=e.Validate();err!=nil{return err}
 if e.SessionID!=session||e.SenderID!=fingerprint{return ErrUnauthorized}
 r.mu.RLock()
 defer r.mu.RUnlock()
 if !r.consent.Allowed(session,fingerprint,e.RecipientID){return ErrConsentDenied}
 peers:=r.destinations[session]
 if peers==nil{return ErrPeerOffline}
 queue,ok:=peers[e.RecipientID]
 if !ok{return ErrPeerOffline}
 select{
 case queue<-e:return nil
 default:return ErrConnectionLimit
 }
}

// CloseSession removes all destinations and discards buffered signaling.
// It does not close channels because consumers may still be reading them.
// Transport writers MUST recheck consent immediately before delivery.
func (r *Router)CloseSession(sessionID string){
 r.mu.Lock()
 defer r.mu.Unlock()
 peers:=r.destinations[sessionID]
 delete(r.destinations,sessionID)
 for _,queue:=range peers{
  for {
   select {
   case <-queue:
   default:
    goto nextQueue
   }
  }
 nextQueue:
 }
}
