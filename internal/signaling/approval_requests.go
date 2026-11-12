package signaling

import (
 "errors"
 "sync"
 "time"
)

var (
 ErrApprovalRequest = errors.New("approval request unavailable")
 ErrApprovalPending = errors.New("approval request already pending")
)
type ApprovalRequest struct {
 SessionID string
 HostFingerprint string
 TechnicianFingerprint string
 ExpiresAt time.Time
}
type pendingApproval struct {
 request ApprovalRequest
 hostConnectionID string
 techConnectionID string
}

// ApprovalRequests models the host-facing request/decision workflow.
// It is internal: a trusted caller must establish host/technician roles and
// deliver the request to a real host UI on the verified host connection.
type ApprovalRequests struct {
 mu sync.Mutex
 pending map[string]pendingApproval
 max int
 now func()time.Time
 coordinator *Coordinator
 connections *Connections
}
func NewApprovalRequests(max int,connections *Connections,coordinator *Coordinator)(*ApprovalRequests,error){
 if max<=0||connections==nil||coordinator==nil{return nil,ErrConsentInvalid}
 return &ApprovalRequests{pending:make(map[string]pendingApproval),max:max,now:time.Now,connections:connections,coordinator:coordinator},nil
}
func (a *ApprovalRequests)Request(sessionID,hostConnectionID,techConnectionID string,ttl time.Duration)(ApprovalRequest,error){
 if ttl<=0||ttl>15*time.Minute{return ApprovalRequest{},ErrApprovalRequest}
 hostSession,host,err:=a.connections.Identity(hostConnectionID)
 if err!=nil{return ApprovalRequest{},err}
 techSession,tech,err:=a.connections.Identity(techConnectionID)
 if err!=nil{return ApprovalRequest{},err}
 if sessionID==""||hostSession!=sessionID||techSession!=sessionID||host==tech{return ApprovalRequest{},ErrApprovalRequest}
 a.mu.Lock()
 defer a.mu.Unlock()
 for id,p:=range a.pending{if !a.now().Before(p.request.ExpiresAt){delete(a.pending,id)}}
 if _,exists:=a.pending[sessionID];exists{return ApprovalRequest{},ErrApprovalPending}
 if len(a.pending)>=a.max{return ApprovalRequest{},ErrConnectionLimit}
 req:=ApprovalRequest{SessionID:sessionID,HostFingerprint:host,TechnicianFingerprint:tech,ExpiresAt:a.now().Add(ttl)}
 a.pending[sessionID]=pendingApproval{request:req,hostConnectionID:hostConnectionID,techConnectionID:techConnectionID}
 return req,nil
}
// Decide must be invoked ONLY by the authenticated host-side UI action.
// Denial clears the request without enrolling or approving a session.
func (a *ApprovalRequests)Decide(sessionID,hostConnectionID string,approve bool)error{
 a.mu.Lock()
 defer a.mu.Unlock()
 pending,ok:=a.pending[sessionID]
 if !ok||!a.now().Before(pending.request.ExpiresAt)||pending.hostConnectionID!=hostConnectionID{return ErrApprovalRequest}
 hostSession,host,err:=a.connections.Identity(hostConnectionID)
 if err!=nil{return err}
 if hostSession!=sessionID||host!=pending.request.HostFingerprint{return ErrApprovalRequest}
 if !approve{delete(a.pending,sessionID);return nil}
 techSession,tech,err:=a.connections.Identity(pending.techConnectionID)
 if err!=nil{return err}
 if techSession!=sessionID||tech!=pending.request.TechnicianFingerprint{return ErrApprovalRequest}
 if err:=a.coordinator.Enroll(sessionID,hostConnectionID,pending.techConnectionID,pending.request.ExpiresAt);err!=nil{return err}
 if err:=a.coordinator.ApproveHost(sessionID,hostConnectionID,pending.techConnectionID);err!=nil{
  a.coordinator.consent.Close(sessionID)
  return err
 }
 delete(a.pending,sessionID)
 return nil
}
func (a *ApprovalRequests)Cancel(sessionID string){
 a.mu.Lock();defer a.mu.Unlock()
 delete(a.pending,sessionID)
}
