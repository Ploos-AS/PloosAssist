package signaling

import (
 "errors"
 "testing"
 "time"
)
func setupApprovalEvents(t *testing.T)(*ApprovalEvents,*ConsentPolicy,ApprovalTicket){
 t.Helper()
 connections,_:=NewConnections(3,time.Minute)
 consent,_:=NewConsentPolicy(2)
 coordinator,_:=NewCoordinator(connections,consent)
 approvals,_:=NewApprovalRequests(2,connections,coordinator)
 events,_:=NewApprovalEvents(2,approvals)
 _=connections.Register("host","s","host-fp")
 _=connections.Register("tech","s","tech-fp")
 _=connections.Register("other","s","other-fp")
 req,err:=approvals.Request("s","host","tech",time.Minute)
 if err!=nil{t.Fatal(err)}
 ticket,err:=events.Issue(req,"host")
 if err!=nil{t.Fatal(err)}
 return events,consent,ticket
}
func TestApprovalEventBoundToHostAndSingleUse(t *testing.T){
 events,consent,ticket:=setupApprovalEvents(t)
 if err:=events.Decide(ticket.ID,"other",true);!errors.Is(err,ErrApprovalEvent){t.Fatalf("other host: %v",err)}
 if err:=events.Decide(ticket.ID,"host",true);err!=nil{t.Fatal(err)}
 if !consent.Allowed("s","host-fp","tech-fp"){t.Fatal("approval not granted")}
 if err:=events.Decide(ticket.ID,"host",true);!errors.Is(err,ErrApprovalEvent){t.Fatalf("replay: %v",err)}
}
func TestApprovalEventDenial(t *testing.T){
 events,consent,ticket:=setupApprovalEvents(t)
 if err:=events.Decide(ticket.ID,"host",false);err!=nil{t.Fatal(err)}
 if consent.Allowed("s","host-fp","tech-fp"){t.Fatal("denial granted consent")}
}
func TestApprovalEventExpiry(t *testing.T){
 events,_,ticket:=setupApprovalEvents(t)
 events.now=func()time.Time{return ticket.ExpiresAt.Add(time.Second)}
 if err:=events.Decide(ticket.ID,"host",true);!errors.Is(err,ErrApprovalEvent){t.Fatalf("expired: %v",err)}
}
