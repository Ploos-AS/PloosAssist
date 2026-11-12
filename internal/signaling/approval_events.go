package signaling

import (
 "crypto/rand"
 "encoding/base64"
 "errors"
 "sync"
 "time"
)

var ErrApprovalEvent = errors.New("invalid or expired approval event")

// ApprovalTicket is an opaque, one-use reference to a pending host UI prompt.
// The ticket is not an authorization capability: only the authenticated host
// connection originally associated with it can submit a decision.
type ApprovalTicket struct {
 ID string
 SessionID string
 TechnicianFingerprint string
 ExpiresAt time.Time
}
type approvalTicketRecord struct {
 ticket ApprovalTicket
 hostConnectionID string
}
type ApprovalEvents struct {
 mu sync.Mutex
 pending map[string]approvalTicketRecord
 now func()time.Time
 max int
 approvals *ApprovalRequests
}
func NewApprovalEvents(max int, approvals *ApprovalRequests)(*ApprovalEvents,error){
 if max<=0||approvals==nil{return nil,ErrConsentInvalid}
 return &ApprovalEvents{pending:make(map[string]approvalTicketRecord),now:time.Now,max:max,approvals:approvals},nil
}

// Issue is invoked only after a trusted session orchestrator has created
// the approval request and selected its actual authenticated host socket.
func (e *ApprovalEvents)Issue(request ApprovalRequest,hostConnectionID string)(ApprovalTicket,error){
 if request.SessionID==""||request.TechnicianFingerprint==""||hostConnectionID==""||!e.now().Before(request.ExpiresAt){return ApprovalTicket{},ErrApprovalEvent}
 e.approvals.mu.Lock()
 pending,ok:=e.approvals.pending[request.SessionID]
 valid:=ok&&pending.hostConnectionID==hostConnectionID&&pending.request.TechnicianFingerprint==request.TechnicianFingerprint&&pending.request.ExpiresAt.Equal(request.ExpiresAt)
 e.approvals.mu.Unlock()
 if !valid{return ApprovalTicket{},ErrApprovalEvent}
 var nonce [32]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{return ApprovalTicket{},err}
 ticket:=ApprovalTicket{ID:base64.RawURLEncoding.EncodeToString(nonce[:]),SessionID:request.SessionID,TechnicianFingerprint:request.TechnicianFingerprint,ExpiresAt:request.ExpiresAt}
 e.mu.Lock();defer e.mu.Unlock()
 for id,r:=range e.pending{if !e.now().Before(r.ticket.ExpiresAt){delete(e.pending,id)}}
 if len(e.pending)>=e.max{return ApprovalTicket{},ErrConnectionLimit}
 e.pending[ticket.ID]=approvalTicketRecord{ticket:ticket,hostConnectionID:hostConnectionID}
 return ticket,nil
}

// Decide consumes the ticket even when approval fails, preventing replay.
// The hostConnectionID MUST be derived from the actual authenticated socket,
// never from a JSON field provided by the client.
func (e *ApprovalEvents)Decide(ticketID,hostConnectionID string,approve bool)error{
 e.mu.Lock()
 record,ok:=e.pending[ticketID]
 if !ok||!e.now().Before(record.ticket.ExpiresAt)||record.hostConnectionID!=hostConnectionID{
  e.mu.Unlock()
  return ErrApprovalEvent
 }
 delete(e.pending,ticketID)
 e.mu.Unlock()
 return e.approvals.Decide(record.ticket.SessionID,hostConnectionID,approve)
}
func (e *ApprovalEvents)Cancel(ticketID string){
 e.mu.Lock();defer e.mu.Unlock()
 delete(e.pending,ticketID)
}

// CancelSession invalidates every outstanding one-use ticket for a session.
func (e *ApprovalEvents)CancelSession(sessionID string){
 e.mu.Lock();defer e.mu.Unlock()
 for id,record:=range e.pending{
  if record.ticket.SessionID==sessionID{delete(e.pending,id)}
 }
}
