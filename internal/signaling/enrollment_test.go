package signaling

import (
 "errors"
 "testing"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/session"
)
func TestCapabilityBackedEnrollment(t *testing.T){
 store,_:=session.NewStore(5*time.Minute)
 id,hostCap,err:=store.Create()
 if err!=nil{t.Fatal(err)}
 techCap,err:=store.Request(id)
 if err!=nil{t.Fatal(err)}
 connections,_:=NewConnections(4,time.Minute)
 roles,_:=NewRoles(2,connections)
 enrollment,_:=NewEnrollment(store,connections,roles)
 _=connections.Register("host",id,"host-fp")
 _=connections.Register("tech",id,"tech-fp")
 _=connections.Register("attacker",id,"attacker-fp")
 _=connections.Register("other","different","other-fp")
 if err:=enrollment.Bind(id,"attacker",HostRole,techCap);!errors.Is(err,ErrEnrollment){t.Fatalf("technician capability as host: %v",err)}
 if err:=enrollment.Bind(id,"attacker",TechnicianRole,hostCap);!errors.Is(err,ErrEnrollment){t.Fatalf("host capability as technician: %v",err)}
 if err:=enrollment.Bind(id,"other",HostRole,hostCap);!errors.Is(err,ErrEnrollment){t.Fatalf("wrong session: %v",err)}
 if err:=enrollment.Bind(id,"host",HostRole,hostCap);err!=nil{t.Fatal(err)}
 if err:=enrollment.Bind(id,"tech",TechnicianRole,techCap);err!=nil{t.Fatal(err)}
 host,tech,err:=roles.Pair(id)
 if err!=nil||host!="host"||tech!="tech"{t.Fatalf("pair: %s %s %v",host,tech,err)}
 roles.Disconnect("host")
 if _,_,err:=roles.Pair(id);!errors.Is(err,ErrRoleBinding){t.Fatalf("disconnect: %v",err)}
}
func TestEnrollmentDeniedAfterPairingRevoke(t *testing.T){
 store,_:=session.NewStore(time.Minute)
 id,hostCap,_:=store.Create()
 techCap,_:=store.Request(id)
 if err:=store.Revoke(id,hostCap);err!=nil{t.Fatal(err)}
 connections,_:=NewConnections(2,time.Minute)
 roles,_:=NewRoles(2,connections)
 enrollment,_:=NewEnrollment(store,connections,roles)
 _=connections.Register("tech",id,"tech-fp")
 if err:=enrollment.Bind(id,"tech",TechnicianRole,techCap);!errors.Is(err,ErrEnrollment){t.Fatalf("revoked: %v",err)}
}
