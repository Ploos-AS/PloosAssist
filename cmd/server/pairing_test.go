package main

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

func call(t *testing.T,h http.Handler,method,path,token string)(int,map[string]any){
 t.Helper()
 r:=httptest.NewRequest(method,path,nil)
 r.RemoteAddr="127.0.0.1:40000"
 if token!=""{r.Header.Set("Authorization","Bearer "+token)}
 w:=httptest.NewRecorder();h.ServeHTTP(w,r)
 var v map[string]any
 if err:=json.Unmarshal(w.Body.Bytes(),&v);err!=nil{t.Fatal(err)}
 return w.Code,v
}
func TestPairingAPI(t *testing.T){
 h:=handler()
 status,created:=call(t,h,"POST","/v1/pairings","")
 if status!=201{t.Fatal(status)}
 id:=created["id"].(string);host:=created["host_capability"].(string)
 path:="/v1/pairings/"+id
 status,requested:=call(t,h,"POST",path+"/request","")
 if status!=200{t.Fatal(status)}
 tech:=requested["technician_capability"].(string)
 if status,_=call(t,h,"GET",path+"/authorization",tech);status!=403{t.Fatal("authorized before consent")}
 if status,_=call(t,h,"POST",path+"/approve","wrong");status!=403{t.Fatal("accepted wrong host capability")}
 if status,_=call(t,h,"POST",path+"/approve",host);status!=200{t.Fatal("host approval failed")}
 if status,_=call(t,h,"GET",path+"/authorization",tech);status!=200{t.Fatal("valid technician rejected")}
 if status,_=call(t,h,"POST",path+"/revoke",host);status!=200{t.Fatal("revoke failed")}
 if status,_=call(t,h,"GET",path+"/authorization",tech);status!=403{t.Fatal("authorized after revoke")}
}
func TestNonLocalRejected(t *testing.T){
 r:=httptest.NewRequest("POST","/v1/pairings",strings.NewReader(""))
 r.RemoteAddr="192.0.2.10:12345"
 w:=httptest.NewRecorder()
 handler().ServeHTTP(w,r)
 if w.Code!=403{t.Fatalf("got %d",w.Code)}
}
