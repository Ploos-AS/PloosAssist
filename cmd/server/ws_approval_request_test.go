package main

import (
 "bytes"
 "context"
 "encoding/json"
 "net/http/httptest"
 "strings"
 "testing"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/hostclient"
 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/gorilla/websocket"
)
func TestConsentRequestDeliveredOnlyToHostSocket(t *testing.T){
 server:=httptest.NewServer(handler())
 defer server.Close()
 resp,err:=server.Client().Post(server.URL+"/v1/pairings","application/json",bytes.NewReader(nil))
 if err!=nil{t.Fatal(err)}
 var pair struct{ID string `json:"id"`;Host string `json:"host_capability"`}
 if err:=json.NewDecoder(resp.Body).Decode(&pair);err!=nil{t.Fatal(err)}
 resp.Body.Close()
 resp,err=server.Client().Post(server.URL+"/v1/pairings/"+pair.ID+"/request","application/json",bytes.NewReader(nil))
 if err!=nil{t.Fatal(err)}
 var tech struct{Cap string `json:"technician_capability"`}
 if err:=json.NewDecoder(resp.Body).Decode(&tech);err!=nil{t.Fatal(err)}
 resp.Body.Close()
 endpoint:="ws"+strings.TrimPrefix(server.URL,"http")+"/v1/ws"
 connect:=func(role,cap string)*websocket.Conn{
  t.Helper()
  key,err:=identity.Generate()
  if err!=nil{t.Fatal(err)}
  ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
  defer cancel()
  conn,_,err:=hostclient.Connect(ctx,endpoint,pair.ID,key)
  if err!=nil{t.Fatal(err)}
  msg,_:=json.Marshal(map[string]string{"type":"role_enroll","role":role,"capability":cap})
  if err:=conn.WriteMessage(websocket.TextMessage,msg);err!=nil{t.Fatal(err)}
  _=conn.SetReadDeadline(time.Now().Add(3*time.Second))
  _,payload,err:=conn.ReadMessage()
  if err!=nil||!bytes.Contains(payload,[]byte("role_enrolled")){t.Fatalf("role enrollment: %s %v",payload,err)}
  return conn
 }
 host:=connect("host",pair.Host)
 defer host.Close()
 technician:=connect("technician",tech.Cap)
 defer technician.Close()
 if err:=host.WriteMessage(websocket.TextMessage,[]byte(`{"type":"approval_check"}`));err!=nil{t.Fatal(err)}
 _,payload,err:=host.ReadMessage()
 if err!=nil||!bytes.Contains(payload,[]byte("approval_none")){t.Fatalf("initial check: %s %v",payload,err)}
 if err:=technician.WriteMessage(websocket.TextMessage,[]byte(`{"type":"approval_request"}`));err!=nil{t.Fatal(err)}
 _,payload,err=technician.ReadMessage()
 if err!=nil||!bytes.Contains(payload,[]byte("approval_requested")){t.Fatalf("request: %s %v",payload,err)}
 if err:=host.WriteMessage(websocket.TextMessage,[]byte(`{"type":"approval_check"}`));err!=nil{t.Fatal(err)}
 _,payload,err=host.ReadMessage()
 if err!=nil{t.Fatal(err)}
 var pending struct{Type string `json:"type"`;TicketID string `json:"ticket_id"`;Fingerprint string `json:"technician_fingerprint"`}
 if err:=json.Unmarshal(payload,&pending);err!=nil||pending.Type!="approval_pending"||pending.TicketID==""||pending.Fingerprint==""{t.Fatalf("pending: %s %v",payload,err)}
 // Receiving a prompt is not consent; signaling is still forbidden.
 if err:=technician.WriteMessage(websocket.TextMessage,[]byte(`{"type":"offer"}`));err!=nil{t.Fatal(err)}
 if _,_,err:=technician.ReadMessage();err==nil{t.Fatal("signaling accepted without approval")}
}
