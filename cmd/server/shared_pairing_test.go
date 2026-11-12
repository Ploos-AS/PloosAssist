package main

import (
 "testing"

 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/Ploos-AS/PloosAssist/internal/signaling"
)
func TestSharedPairingStoreSupportsVerifiedRoleEnrollment(t *testing.T){
 pairing:=newPairingAPI()
 ws:=newWSAPI()
 ws.enrollment=mustEnrollment(pairing.store,ws.connections,ws.roles)
 id,hostCap,err:=pairing.store.Create()
 if err!=nil{t.Fatal(err)}
 techCap,err:=pairing.store.Request(id)
 if err!=nil{t.Fatal(err)}
 hostKey,err:=identity.Generate()
 if err!=nil{t.Fatal(err)}
 techKey,err:=identity.Generate()
 if err!=nil{t.Fatal(err)}
 hostFP,err:=identity.Fingerprint(hostKey.Public)
 if err!=nil{t.Fatal(err)}
 techFP,err:=identity.Fingerprint(techKey.Public)
 if err!=nil{t.Fatal(err)}
 if err:=ws.connections.Register("host-conn",id,hostFP);err!=nil{t.Fatal(err)}
 if err:=ws.connections.Register("tech-conn",id,techFP);err!=nil{t.Fatal(err)}
 if err:=ws.enrollment.Bind(id,"host-conn",signaling.HostRole,hostCap);err!=nil{t.Fatal(err)}
 if err:=ws.enrollment.Bind(id,"tech-conn",signaling.TechnicianRole,techCap);err!=nil{t.Fatal(err)}
 h,tech,err:=ws.roles.Pair(id)
 if err!=nil||h!="host-conn"||tech!="tech-conn"{t.Fatalf("role pair: %q %q %v",h,tech,err)}
 ws.cleanupConnection("host-conn")
 if _,_,err:=ws.roles.Pair(id);err==nil{t.Fatal("role pair survived host disconnect")}
}
