package main

import (
 "bytes"
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/hostclient"
 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/gorilla/websocket"
)
func TestWebSocketRoleEnrollmentWithRealPairing(t *testing.T){
 server:=httptest.NewServer(handler())
 defer server.Close()
 client:=server.Client()
 resp,err:=client.Post(server.URL+"/v1/pairings","application/json",bytes.NewReader(nil))
 if err!=nil{t.Fatal(err)}
 var pair struct{ID string `json:"id"`;HostCapability string `json:"host_capability"`}
 if err:=json.NewDecoder(resp.Body).Decode(&pair);err!=nil{t.Fatal(err)}
 resp.Body.Close()
 if resp.StatusCode!=http.StatusCreated{t.Fatalf("pairing status %d",resp.StatusCode)}
 resp,err=client.Post(server.URL+"/v1/pairings/"+pair.ID+"/request","application/json",bytes.NewReader(nil))
 if err!=nil{t.Fatal(err)}
 var tech struct{Capability string `json:"technician_capability"`}
 if err:=json.NewDecoder(resp.Body).Decode(&tech);err!=nil{t.Fatal(err)}
 resp.Body.Close()
 if tech.Capability==""{t.Fatal("missing technician capability")}
 endpoint:="ws"+strings.TrimPrefix(server.URL,"http")+"/v1/ws"
 for _,tc:=range []struct{name,role,cap string;accepted bool}{
  {"invalid-host","host","wrong",false},
  {"host","host",pair.HostCapability,true},
  {"technician","technician",tech.Capability,true},
 }{
  t.Run(tc.name,func(t *testing.T){
   key,err:=identity.Generate()
   if err!=nil{t.Fatal(err)}
   ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
   defer cancel()
   conn,_,err:=hostclient.Connect(ctx,endpoint,pair.ID,key)
   if err!=nil{t.Fatal(err)}
   defer conn.Close()
   msg,_:=json.Marshal(map[string]string{"type":"role_enroll","role":tc.role,"capability":tc.cap})
   if err:=conn.WriteMessage(websocket.TextMessage,msg);err!=nil{t.Fatal(err)}
   _=conn.SetReadDeadline(time.Now().Add(3*time.Second))
   _,payload,err:=conn.ReadMessage()
   if tc.accepted{
    if err!=nil{t.Fatalf("enrollment rejected: %v",err)}
    var ack struct{Type string `json:"type"`;Role string `json:"role"`}
    if err:=json.Unmarshal(payload,&ack);err!=nil||ack.Type!="role_enrolled"||ack.Role!=tc.role{t.Fatalf("unexpected ack %s: %v",payload,err)}
    // Enrollment is not approval: signaling must still be rejected.
    if err:=conn.WriteMessage(websocket.TextMessage,[]byte(`{"type":"offer"}`));err!=nil{t.Fatal(err)}
    if _,_,err:=conn.ReadMessage();err==nil{t.Fatal("signaling accepted without host consent")}
   }else if err==nil{t.Fatalf("invalid credential accepted: %s",payload)}
  })
 }
}
