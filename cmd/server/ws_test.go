package main

import (
 "testing"
 "net/http"
 "net/http/httptest"
 "encoding/base64"

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
}

func TestWSRejectsExcessConnectionsBeforeUpgrade(t *testing.T){
 a:=newWSAPI()
 mux:=http.NewServeMux()
 a.routes(mux)
 for i:=0;i<cap(a.slots);i++{a.slots<-struct{}{}}
 defer func(){for i:=0;i<cap(a.slots);i++{<-a.slots}}()
 kp,err:=identity.Generate();if err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest("GET","/v1/ws?session_id=test&public_key="+base64.RawURLEncoding.EncodeToString(kp.Public),nil)
 rec:=httptest.NewRecorder()
 mux.ServeHTTP(rec,req)
 if rec.Code!=http.StatusServiceUnavailable{t.Fatalf("status %d, want 503",rec.Code)}
}
