package signaling

import (
 "errors"
 "testing"
 "time"
)
func TestRoleBindingRequiresVerifiedDistinctPeers(t *testing.T){
 c,_:=NewConnections(4,time.Minute)
 roles,_:=NewRoles(2,c)
 expiry:=time.Now().Add(time.Minute)
 _=c.Register("host","s","host-fp")
 _=c.Register("tech","s","tech-fp")
 _=c.Register("other","other-session","other-fp")
 if err:=roles.Bind("s","other",HostRole,expiry);!errors.Is(err,ErrRoleBinding){t.Fatalf("wrong session: %v",err)}
 if err:=roles.Bind("s","host",HostRole,expiry);err!=nil{t.Fatal(err)}
 if err:=roles.Bind("s","host",TechnicianRole,expiry);!errors.Is(err,ErrRoleBinding){t.Fatalf("self-pairing: %v",err)}
 if err:=roles.Bind("s","tech",TechnicianRole,expiry);err!=nil{t.Fatal(err)}
 h,tech,err:=roles.Pair("s")
 if err!=nil||h!="host"||tech!="tech"{t.Fatalf("pair: %s %s %v",h,tech,err)}
 if err:=roles.Bind("s","tech",TechnicianRole,expiry);!errors.Is(err,ErrRoleOccupied){t.Fatalf("duplicate role: %v",err)}
 roles.Disconnect("host")
 if _,_,err:=roles.Pair("s");!errors.Is(err,ErrRoleBinding){t.Fatalf("disconnected: %v",err)}
}
func TestRoleBindingExpiredOrDisconnected(t *testing.T){
 c,_:=NewConnections(2,time.Minute)
 roles,_:=NewRoles(2,c)
 _=c.Register("host","s","h")
 _=c.Register("tech","s","t")
 expiry:=time.Now().Add(time.Minute)
 if err:=roles.Bind("s","host",HostRole,expiry);err!=nil{t.Fatal(err)}
 if err:=roles.Bind("s","tech",TechnicianRole,expiry);err!=nil{t.Fatal(err)}
 c.Close("tech")
 if _,_,err:=roles.Pair("s");!errors.Is(err,ErrRoleBinding){t.Fatalf("closed technician: %v",err)}
 roles.Close("s")
 if _,_,err:=roles.Pair("s");!errors.Is(err,ErrRoleBinding){t.Fatalf("closed session: %v",err)}
}
