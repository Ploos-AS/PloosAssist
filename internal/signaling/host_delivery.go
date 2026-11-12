package signaling

import (
 "errors"
 "sync"
 "time"
)

var ErrHostDelivery = errors.New("host approval delivery unavailable")

// HostDelivery holds a pending approval prompt for exactly the authenticated
// host connection selected by trusted orchestration. It does not authorize
// the requester or expose any network endpoint.
type HostDelivery struct {
 mu sync.Mutex
 pending map[string]hostDeliveryRecord
 connections *Connections
 approvals *ApprovalRequests
 events *ApprovalEvents
 max int
 now func()time.Time
}
type hostDeliveryRecord struct {
 hostConnectionID string
 hostFingerprint string
 ticket ApprovalTicket
}
func NewHostDelivery(max int,connections *Connections,approvals *ApprovalRequests,events *ApprovalEvents)(*HostDelivery,error){
 if max<=0||connections==nil||approvals==nil||events==nil{return nil,ErrConsentInvalid}
 return &HostDelivery{pending:make(map[string]hostDeliveryRecord),connections:connections,approvals:approvals,events:events,max:max,now:time.Now},nil
}

// Queue is called by a trusted pairing orchestrator, never by an arbitrary
// client claiming a host role. The returned ticket must be displayed only
// through the host's authenticated connection.
func (d *HostDelivery)Queue(request ApprovalRequest,hostConnectionID string)(ApprovalTicket,error){
 session,fingerprint,err:=d.connections.Identity(hostConnectionID)
 if err!=nil{return ApprovalTicket{},err}
 if session!=request.SessionID||fingerprint!=request.HostFingerprint{return ApprovalTicket{},ErrHostDelivery}
 d.mu.Lock()
 defer d.mu.Unlock()
 for id,rec:=range d.pending{
  if !d.now().Before(rec.ticket.ExpiresAt){
   d.events.Cancel(rec.ticket.ID)
   delete(d.pending,id)
  }
 }
 if _,ok:=d.pending[hostConnectionID];ok{return ApprovalTicket{},ErrApprovalPending}
 if len(d.pending)>=d.max{return ApprovalTicket{},ErrConnectionLimit}
 ticket,err:=d.events.Issue(request,hostConnectionID)
 if err!=nil{return ApprovalTicket{},err}
 d.pending[hostConnectionID]=hostDeliveryRecord{hostConnectionID:hostConnectionID,hostFingerprint:fingerprint,ticket:ticket}
 return ticket,nil
}

// Take verifies the live connection before revealing the pending prompt.
// It is safe to call repeatedly for the same host connection.
func (d *HostDelivery)Take(hostConnectionID string)(ApprovalTicket,error){
 d.mu.Lock();defer d.mu.Unlock()
 rec,ok:=d.pending[hostConnectionID]
 if !ok||!d.now().Before(rec.ticket.ExpiresAt){return ApprovalTicket{},ErrHostDelivery}
 session,fingerprint,err:=d.connections.Identity(hostConnectionID)
 if err!=nil{return ApprovalTicket{},err}
 if session!=rec.ticket.SessionID||fingerprint!=rec.hostFingerprint{return ApprovalTicket{},ErrHostDelivery}
 return rec.ticket,nil
}

// Decide consumes a delivered request only after the host's authenticated
// connection has submitted the matching one-use ticket.
func (d *HostDelivery)Decide(hostConnectionID,ticketID string,approve bool)error{
 d.mu.Lock()
 rec,ok:=d.pending[hostConnectionID]
 if !ok||rec.ticket.ID!=ticketID||!d.now().Before(rec.ticket.ExpiresAt){
  d.mu.Unlock()
  return ErrHostDelivery
 }
 session,fingerprint,err:=d.connections.Identity(hostConnectionID)
 if err!=nil||session!=rec.ticket.SessionID||fingerprint!=rec.hostFingerprint{
  d.mu.Unlock()
  return ErrHostDelivery
 }
 delete(d.pending,hostConnectionID)
 d.mu.Unlock()
 return d.events.Decide(ticketID,hostConnectionID,approve)
}
func (d *HostDelivery)Disconnect(hostConnectionID string){
 d.mu.Lock()
 rec,ok:=d.pending[hostConnectionID]
 delete(d.pending,hostConnectionID)
 d.mu.Unlock()
 if ok{
  d.events.Cancel(rec.ticket.ID)
  d.approvals.Cancel(rec.ticket.SessionID)
 }
}

// CancelSession removes all pending host prompts for this session.
func (d *HostDelivery)CancelSession(sessionID string){
 d.mu.Lock()
 defer d.mu.Unlock()
 for id,record:=range d.pending{
  if record.ticket.SessionID==sessionID{
   d.events.Cancel(record.ticket.ID)
   delete(d.pending,id)
  }
 }
 d.approvals.Cancel(sessionID)
}
