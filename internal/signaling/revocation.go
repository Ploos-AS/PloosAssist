package signaling

import (
 "errors"
 "sync"

 "github.com/Ploos-AS/PloosAssist/internal/session"
)

var ErrRevocationSetup=errors.New("revocation dependencies unavailable")

// Revocation coordinates in-process session shutdown after host-capability
// authorization. It is not yet wired to any transport or media channel.
type Revocation struct {
 mu sync.Mutex
 pairing *session.Store
 consent *ConsentPolicy
 roles *Roles
 approvals *ApprovalRequests
 events *ApprovalEvents
 delivery *HostDelivery
 router *Router
}
func NewRevocation(pairing *session.Store,consent *ConsentPolicy,roles *Roles,approvals *ApprovalRequests,events *ApprovalEvents,delivery *HostDelivery,router *Router)(*Revocation,error){
 if pairing==nil||consent==nil||roles==nil||approvals==nil||events==nil||delivery==nil||router==nil{return nil,ErrRevocationSetup}
 return &Revocation{pairing:pairing,consent:consent,roles:roles,approvals:approvals,events:events,delivery:delivery,router:router},nil
}

// Revoke requires the independent host pairing capability. Once authorized,
// no new internal routing operation can pass consent checks; pending prompts,
// role bindings and buffered signaling are then discarded. Callers must also
// terminate real transports and media streams when those are implemented.
func (r *Revocation)Revoke(sessionID,hostCapability string)error{
 r.mu.Lock();defer r.mu.Unlock()
 if err:=r.pairing.Revoke(sessionID,hostCapability);err!=nil{return err}
 r.consent.Close(sessionID)
 r.roles.Close(sessionID)
 r.delivery.CancelSession(sessionID)
 r.events.CancelSession(sessionID)
 r.approvals.Cancel(sessionID)
 r.router.CloseSession(sessionID)
 return nil
}
