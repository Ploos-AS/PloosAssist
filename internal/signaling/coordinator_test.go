package signaling

import (
 "errors"
 "testing"
 "time"
)
func TestCoordinatorRequiresLiveVerifiedHostAndExactTechnician(t *testing.T){
 connections,_:=NewConnections(4,time.Minute)
 consent,_:=NewConsentPolicy(2)
 coordinator,_:=NewCoordinator(connections,consent)
 _=connections.Register("host","s","host-fp")
 _=connections.Register("tech","s","tech-fp")
 _=connections.Register("stranger","other","stranger-fp")
 expiry:=time.Now().Add(time.Minute)
 if err:=coordinator.Enroll("s","host","tech",expiry);err!=nil{t.Fatal(err)}
 if err:=coordinator.ApproveHost("s","tech","host");!errors.Is(err,ErrConsentDenied){t.Fatalf("technician cannot approve: %v",err)}
 if err:=coordinator.ApproveHost("s","host","stranger");!errors.Is(err,ErrHostIdentity){t.Fatalf("wrong session: %v",err)}
 if consent.Allowed("s","host-fp","tech-fp"){t.Fatal("consent before host approval")}
 if err:=coordinator.ApproveHost("s","host","tech");err!=nil{t.Fatal(err)}
 if !consent.Allowed("s","host-fp","tech-fp"){t.Fatal("approved pair blocked")}
 if err:=coordinator.RevokeHost("s","tech");!errors.Is(err,ErrConsentDenied){t.Fatalf("technician revoked host: %v",err)}
 if err:=coordinator.RevokeHost("s","host");err!=nil{t.Fatal(err)}
 if consent.Allowed("s","host-fp","tech-fp"){t.Fatal("revocation ignored")}
}
func TestCoordinatorDisconnectedHostCannotApprove(t *testing.T){
 connections,_:=NewConnections(2,time.Minute)
 consent,_:=NewConsentPolicy(1)
 coordinator,_:=NewCoordinator(connections,consent)
 _=connections.Register("host","s","h")
 _=connections.Register("tech","s","t")
 if err:=coordinator.Enroll("s","host","tech",time.Now().Add(time.Minute));err!=nil{t.Fatal(err)}
 connections.Close("host")
 if err:=coordinator.ApproveHost("s","host","tech");!errors.Is(err,ErrConnectionMissing){t.Fatalf("disconnected host: %v",err)}
 if consent.Allowed("s","h","t"){t.Fatal("approved disconnected host")}
}
