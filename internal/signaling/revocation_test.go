package signaling

import (
 "errors"
 "testing"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/session"
)
func TestRevocationBlocksRoutingAndClearsRoles(t *testing.T){
 store,_:=session.NewStore(time.Minute)
 id,hostCap,err:=store.Create()
 if err!=nil{t.Fatal(err)}
 techCap,err:=store.Request(id)
 if err!=nil{t.Fatal(err)}
 connections,_:=NewConnections(4,time.Minute)
 _=connections.Register("host",id,"host-fp")
 _=connections.Register("tech",id,"tech-fp")
 roles,_:=NewRoles(4,connections)
 enrollment,_:=NewEnrollment(store,connections,roles)
 if err:=enrollment.Bind(id,"host",HostRole,hostCap);err!=nil{t.Fatal(err)}
 if err:=enrollment.Bind(id,"tech",TechnicianRole,techCap);err!=nil{t.Fatal(err)}
 consent,_:=NewConsentPolicy(4)
 coordinator,_:=NewCoordinator(connections,consent)
 approvals,_:=NewApprovalRequests(4,connections,coordinator)
 events,_:=NewApprovalEvents(4,approvals)
 delivery,_:=NewHostDelivery(4,connections,approvals,events)
 router,_:=NewRouter(connections,consent)
 revocation,_:=NewRevocation(store,consent,roles,approvals,events,delivery,router)
 req,err:=approvals.Request(id,"host","tech",time.Minute)
 if err!=nil{t.Fatal(err)}
 ticket,err:=delivery.Queue(req,"host")
 if err!=nil{t.Fatal(err)}
 if err:=delivery.Decide("host",ticket.ID,true);err!=nil{t.Fatal(err)}
 if !consent.Allowed(id,"host-fp","tech-fp"){t.Fatal("consent not active")}
 if _,err:=router.Attach("tech");err!=nil{t.Fatal(err)}
 if err:=revocation.Revoke(id,"incorrect");!errors.Is(err,session.ErrUnauthorized){t.Fatalf("unauthorized revoke: %v",err)}
 if !consent.Allowed(id,"host-fp","tech-fp"){t.Fatal("unauthorized revoke changed consent")}
 if err:=revocation.Revoke(id,hostCap);err!=nil{t.Fatal(err)}
 if consent.Allowed(id,"host-fp","tech-fp"){t.Fatal("consent survived revoke")}
 if _,_,err:=roles.Pair(id);!errors.Is(err,ErrRoleBinding){t.Fatalf("role survived: %v",err)}
 if _,err:=delivery.Take("host");!errors.Is(err,ErrHostDelivery){t.Fatalf("prompt survived: %v",err)}
 if err:=enrollment.Bind(id,"host",HostRole,hostCap);!errors.Is(err,ErrEnrollment){t.Fatalf("re-enrolled revoked session: %v",err)}
}
