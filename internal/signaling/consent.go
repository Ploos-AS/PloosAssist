package signaling

import (
 "errors"
 "sync"
 "time"
)

var (
 ErrConsentDenied = errors.New("consent denied")
 ErrConsentExists = errors.New("consent session already exists")
 ErrConsentInvalid = errors.New("invalid consent session")
)

type consentSession struct {
 host string
 technician string
 expires time.Time
 approved bool
}

// ConsentPolicy records approval of a specific verified technician identity
// by the host of a specific session. All methods are transport-independent.
type ConsentPolicy struct {
 mu sync.Mutex
 sessions map[string]consentSession
 now func()time.Time
 max int
}

func NewConsentPolicy(max int)(*ConsentPolicy,error){
 if max<=0{return nil,ErrConsentInvalid}
 return &ConsentPolicy{sessions:make(map[string]consentSession),now:time.Now,max:max},nil
}

// Create must be invoked only by trusted session orchestration after the
// host and technician have independently proven key possession.
func (p *ConsentPolicy)Create(sessionID,hostFingerprint,technicianFingerprint string,expires time.Time)error{
 p.mu.Lock();defer p.mu.Unlock()
 if sessionID==""||hostFingerprint==""||technicianFingerprint==""||
  hostFingerprint==technicianFingerprint||!expires.After(p.now()){return ErrConsentInvalid}
 p.prune()
 if _,ok:=p.sessions[sessionID];ok{return ErrConsentExists}
 if len(p.sessions)>=p.max{return ErrConnectionLimit}
 p.sessions[sessionID]=consentSession{host:hostFingerprint,technician:technicianFingerprint,expires:expires}
 return nil
}

// Approve is callable only after host consent has been verified out-of-band.
// A host fingerprint supplied by an untrusted request is NOT authentication.
func (p *ConsentPolicy)Approve(sessionID,authenticatedHost,technicianFingerprint string)error{
 p.mu.Lock();defer p.mu.Unlock()
 s,ok:=p.sessions[sessionID]
 if !ok||!p.now().Before(s.expires)||s.host!=authenticatedHost||s.technician!=technicianFingerprint{return ErrConsentDenied}
 s.approved=true
 p.sessions[sessionID]=s
 return nil
}
func (p *ConsentPolicy)Allowed(sessionID,senderFingerprint,recipientFingerprint string)bool{
 p.mu.Lock();defer p.mu.Unlock()
 s,ok:=p.sessions[sessionID]
 if !ok||!s.approved||!p.now().Before(s.expires){return false}
 return (senderFingerprint==s.host&&recipientFingerprint==s.technician)||
  (senderFingerprint==s.technician&&recipientFingerprint==s.host)
}
func (p *ConsentPolicy)Revoke(sessionID,authenticatedHost string)error{
 p.mu.Lock();defer p.mu.Unlock()
 s,ok:=p.sessions[sessionID]
 if !ok||s.host!=authenticatedHost{return ErrConsentDenied}
 s.approved=false
 p.sessions[sessionID]=s
 return nil
}
func (p *ConsentPolicy)Close(sessionID string){
 p.mu.Lock();defer p.mu.Unlock()
 delete(p.sessions,sessionID)
}
func (p *ConsentPolicy)prune(){
 now:=p.now()
 for id,s:=range p.sessions{if !now.Before(s.expires){delete(p.sessions,id)}}
}
