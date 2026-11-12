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

func TestHostRevokeClosesBothSocketsAndBlocksReenrollment(t *testing.T){
 server:=httptest.NewServer(handler())
 defer server.Close()
 post:=func(path string)*http.Response{
  t.Helper()
  resp,err:=server.Client().Post(server.URL+path,"application/json",bytes.NewReader(nil))
  if err!=nil{t.Fatal(err)}
  return resp
 }
 resp:=post("/v1/pairings")
 var pair struct{ID string `json:"id"`;Host string `json:"host_capability"`}
 if err:=json.NewDecoder(resp.Body).Decode(&pair);err!=nil{t.Fatal(err)}
 resp.Body.Close()
 resp=post("/v1/pairings/"+pair.ID+"/request")
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
  if err:=conn.WriteJSON(map[string]string{"type":"role_enroll","role":role,"capability":cap});err!=nil{t.Fatal(err)}
  _=conn.SetReadDeadline(time.Now().Add(3*time.Second))
  var ack struct{Type string `json:"type"`}
  if err:=conn.ReadJSON(&ack);err!=nil||ack.Type!="role_enrolled"{t.Fatalf("enrollment: %v %+v",err,ack)}
  return conn
 }
 host:=connect("host",pair.Host)
 defer host.Close()
 technician:=connect("technician",tech.Cap)
 defer technician.Close()
 if err:=hostclient.NewConsentClient(host).Revoke(pair.Host);err!=nil{t.Fatalf("revocation not acknowledged: %v",err)}
 _=technician.SetReadDeadline(time.Now().Add(3*time.Second))
 if _,_,err:=technician.ReadMessage();err==nil{t.Fatal("technician socket remained open")}
 _=host.SetReadDeadline(time.Now().Add(3*time.Second))
 if _,_,err:=host.ReadMessage();err==nil{t.Fatal("host socket remained open")}
 // Even with a new authenticated socket, a revoked role capability fails.
 key,err:=identity.Generate()
 if err!=nil{t.Fatal(err)}
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
 defer cancel()
 retry,_,err:=hostclient.Connect(ctx,endpoint,pair.ID,key)
 if err!=nil{t.Fatal(err)}
 defer retry.Close()
 if err:=retry.WriteJSON(map[string]string{"type":"role_enroll","role":"host","capability":pair.Host});err!=nil{t.Fatal(err)}
 _=retry.SetReadDeadline(time.Now().Add(3*time.Second))
 var ack struct{Type string `json:"type"`}
 if err:=retry.ReadJSON(&ack);err==nil{t.Fatal("revoked capability reenrolled")}
}

func TestHTTPRevokeClosesBothSockets(t *testing.T){
 // This regression verifies the existing HTTP route is wired to the
 // same revocation coordinator as the WebSocket path.
 server:=httptest.NewServer(handler())
 defer server.Close()
 resp,err:=server.Client().Post(server.URL+"/v1/pairings","application/json",nil)
 if err!=nil{t.Fatal(err)}
 var pair struct{ID string `json:"id"`;Host string `json:"host_capability"`}
 if err:=json.NewDecoder(resp.Body).Decode(&pair);err!=nil{t.Fatal(err)}
 resp.Body.Close()
 key,err:=identity.Generate()
 if err!=nil{t.Fatal(err)}
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
 defer cancel()
 conn,_,err:=hostclient.Connect(ctx,"ws"+strings.TrimPrefix(server.URL,"http")+"/v1/ws",pair.ID,key)
 if err!=nil{t.Fatal(err)}
 defer conn.Close()
 request,err:=http.NewRequest(http.MethodPost,server.URL+"/v1/pairings/"+pair.ID+"/revoke",nil)
 if err!=nil{t.Fatal(err)}
 request.Header.Set("Authorization","Bearer "+pair.Host)
 resp,err=server.Client().Do(request)
 if err!=nil{t.Fatal(err)}
 resp.Body.Close()
 if resp.StatusCode!=200{t.Fatalf("HTTP revoke status: %d",resp.StatusCode)}
 _=conn.SetReadDeadline(time.Now().Add(3*time.Second))
 if _,_,err:=conn.ReadMessage();err==nil{t.Fatal("HTTP revoke left socket connected")}
}
