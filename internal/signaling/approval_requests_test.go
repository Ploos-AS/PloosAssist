package signaling

import (
 "errors"
 "testing"
 "time"
)
func TestHostApprovalWorkflow(t *testing.T){
 connections,_:=NewConnections(4,time.Minute)
 consent,_:=NewConsentPolicy(4)
 coordinator,_:=NewCoordinator(connections,consent)
 approvals,_:=NewApprovalRequests(4,connections,coordinator)
 _=connections.Register("host","s","host-fp")
 _=connections.Register("tech","s","tech-fp")
 _=connections.Register("stranger","s","stranger-fp")
 req,err:=approvals.Request("s","host","tech",time.Minute)
 if err!=nil||req.TechnicianFingerprint!="tech-fp"{t.Fatalf("request: %+v %v",req,err)}
 if err:=approvals.Decide("s","tech",true);!errors.Is(err,ErrApprovalRequest){t.Fatalf("technician approval: %v",err)}
 if err:=approvals.Decide("s","stranger",true);!errors.Is(err,ErrApprovalRequest){t.Fatalf("stranger approval: %v",err)}
 if consent.Allowed("s","host-fp","tech-fp"){t.Fatal("premature consent")}
 if err:=approvals.Decide("s","host",true);err!=nil{t.Fatal(err)}
 if !consent.Allowed("s","host-fp","tech-fp"){t.Fatal("approval missing")}
 if err:=coordinator.RevokeHost("s","host");err!=nil{t.Fatal(err)}
 if consent.Allowed("s","host-fp","tech-fp"){t.Fatal("revocation ignored")}
}
func TestHostDenialAndDisconnectedTechnician(t *testing.T){
 connections,_:=NewConnections(2,time.Minute)
 consent,_:=NewConsentPolicy(2)
 coordinator,_:=NewCoordinator(connections,consent)
 approvals,_:=NewApprovalRequests(2,connections,coordinator)
 _=connections.Register("host","s","h")
 _=connections.Register("tech","s","t")
 _,err:=approvals.Request("s","host","tech",time.Minute)
 if err!=nil{t.Fatal(err)}
 if err:=approvals.Decide("s","host",false);err!=nil{t.Fatal(err)}
 if consent.Allowed("s","h","t"){t.Fatal("denial allowed signaling")}
 _,err=approvals.Request("s","host","tech",time.Minute)
 if err!=nil{t.Fatal(err)}
 connections.Close("tech")
 if err:=approvals.Decide("s","host",true);!errors.Is(err,ErrConnectionMissing){t.Fatalf("disconnected technician: %v",err)}
 if consent.Allowed("s","h","t"){t.Fatal("disconnected technician approved")}
}
func TestApprovalRequestExpiry(t *testing.T){
 connections,_:=NewConnections(2,time.Minute)
 consent,_:=NewConsentPolicy(2)
 coordinator,_:=NewCoordinator(connections,consent)
 approvals,_:=NewApprovalRequests(2,connections,coordinator)
 _=connections.Register("host","s","h")
 _=connections.Register("tech","s","t")
 _,err:=approvals.Request("s","host","tech",time.Millisecond)
 if err!=nil{t.Fatal(err)}
 time.Sleep(3*time.Millisecond)
 if err:=approvals.Decide("s","host",true);!errors.Is(err,ErrApprovalRequest){t.Fatalf("expired: %v",err)}
}
