package main

import (
 "testing"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/identity"
)

func TestWSAPILifecycleStartsEmpty(t *testing.T){
 a:=newWSAPI()
 if n:=a.connections.Count();n!=0{t.Fatalf("unexpected connections: %d",n)}
 kp,err:=identity.Generate();if err!=nil{t.Fatal(err)}
 id,err:=a.auth.Open("test-session",kp.Public);if err!=nil{t.Fatal(err)}
 defer a.auth.Close(id)
 challengeID,ch,err:=a.auth.Challenge(id);if err!=nil{t.Fatal(err)}
 sig,err:=identity.Sign(kp.Private,ch);if err!=nil{t.Fatal(err)}
 fp,err:=a.auth.Authenticate(id,challengeID,sig);if err!=nil{t.Fatal(err)}
 session,verifiedFP,err:=a.auth.Identity(id)
 if err!=nil||session!="test-session"||verifiedFP!=fp{t.Fatalf("unexpected verified identity: %v",err)}
 if err:=a.connections.Register(id,session,verifiedFP);err!=nil{t.Fatal(err)}
 if n:=a.connections.Count();n!=1{t.Fatalf("registered count: %d",n)}
 a.connections.Close(id)
 if n:=a.connections.Count();n!=0{t.Fatalf("connection leaked: %d",n)}
 _=time.Second
}
