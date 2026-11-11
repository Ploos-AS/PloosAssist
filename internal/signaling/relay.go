package signaling

import (
 "errors"
 "sync"
)

// Relay is a transport-independent authorization boundary for signaling.
// The caller MUST authenticate both peers and obtain consent before Register.
type Relay struct {
 mu sync.RWMutex
 sessions map[string]*peers
}
type peers struct {
 host string
 technician string
 approved bool
}
var (
 ErrUnauthorized = errors.New("unauthorized signaling")
 ErrSessionExists = errors.New("session already registered")
 ErrInvalidPeers = errors.New("invalid peer identities")
)
func NewRelay()*Relay{return &Relay{sessions:make(map[string]*peers)}}

// Register records the two authenticated identity fingerprints for a session.
// Approval is deliberately separate and defaults to false.
func (r *Relay) Register(sessionID,hostID,technicianID string)error{
 if sessionID==""||hostID==""||technicianID==""||hostID==technicianID{return ErrInvalidPeers}
 r.mu.Lock();defer r.mu.Unlock()
 if _,ok:=r.sessions[sessionID];ok{return ErrSessionExists}
 r.sessions[sessionID]=&peers{host:hostID,technician:technicianID}
 return nil
}
// Approve requires the caller to have already verified consent by the host.
// It must not be called directly from an untrusted network request.
func (r *Relay) Approve(sessionID,hostID string)error{
 r.mu.Lock();defer r.mu.Unlock()
 p,ok:=r.sessions[sessionID]
 if !ok||p.host!=hostID{return ErrUnauthorized}
 p.approved=true
 return nil
}
func (r *Relay) Revoke(sessionID,hostID string)error{
 r.mu.Lock();defer r.mu.Unlock()
 p,ok:=r.sessions[sessionID]
 if !ok||p.host!=hostID{return ErrUnauthorized}
 p.approved=false
 return nil
}
func (r *Relay) Remove(sessionID string){r.mu.Lock();defer r.mu.Unlock();delete(r.sessions,sessionID)}

// Authorize derives sender identity from an authenticated transport parameter,
// never from an untrusted JSON sender field alone.
func (r *Relay) Authorize(authenticatedSender string,e Envelope)error{
 if err:=e.Validate();err!=nil{return err}
 r.mu.RLock();defer r.mu.RUnlock()
 p,ok:=r.sessions[e.SessionID]
 if !ok||!p.approved||authenticatedSender!=e.SenderID{return ErrUnauthorized}
 if (e.SenderID==p.host && e.RecipientID==p.technician)||
    (e.SenderID==p.technician && e.RecipientID==p.host){return nil}
 return ErrUnauthorized
}
