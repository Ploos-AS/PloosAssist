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
func TestHostOnlyApprovalDecisionOverWebSocket(t *testing.T){
 for _,approve:=range []bool{false,true}{
  t.Run(map[bool]string{false:"deny",true:"approve"}[approve],func(t *testing.T){
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
    if err!=nil||!bytes.Contains(payload,[]byte("role_enrolled")){t.Fatalf("enrollment: %s %v",payload,err)}
    return conn
   }
   host:=connect("host",pair.Host)
   defer host.Close()
   technician:=connect("technician",tech.Cap)
   defer technician.Close()
   send:=func(conn *websocket.Conn,v any)[]byte{
    t.Helper()
    msg,_:=json.Marshal(v)
    if err:=conn.WriteMessage(websocket.TextMessage,msg);err!=nil{t.Fatal(err)}
    _,reply,err:=conn.ReadMessage()
    if err!=nil{t.Fatal(err)}
    return reply
   }
   send(technician,map[string]any{"type":"approval_request"})
   pending:=send(host,map[string]any{"type":"approval_check"})
   var ticket struct{ID string `json:"ticket_id"`}
   if err:=json.Unmarshal(pending,&ticket);err!=nil||ticket.ID==""{t.Fatalf("ticket: %s %v",pending,err)}
   decision:=send(host,map[string]any{"type":"approval_decide","ticket_id":ticket.ID,"approve":approve})
   var ack struct{Type string `json:"type"`;State string `json:"state"`}
   if err:=json.Unmarshal(decision,&ack);err!=nil||ack.Type!="approval_decided"{t.Fatalf("decision: %s %v",decision,err)}
   expected:="denied"
   if approve{expected="approved"}
   if ack.State!=expected{t.Fatalf("state %q expected %q",ack.State,expected)}
   if err:=host.WriteJSON(map[string]any{"type":"approval_decide","ticket_id":ticket.ID,"approve":true});err!=nil{t.Fatal(err)}
   if _,_,err:=host.ReadMessage();err==nil{t.Fatal("replayed ticket accepted")}
  })
 }
}
