package main

import (
 "context"
 "errors"
 "net/http/httptest"
 "strings"
 "testing"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/hostclient"
 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/gorilla/websocket"
)

// Exercises the actual guarded server handler and real Linux host-client
// handshake, rather than a mock WebSocket endpoint.
func TestHostClientAuthenticatesAgainstRealServer(t *testing.T){
 server:=httptest.NewServer(handler())
 defer server.Close()
 endpoint:="ws"+strings.TrimPrefix(server.URL,"http")+"/v1/ws"
 key,err:=identity.Generate()
 if err!=nil{t.Fatal(err)}
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
 defer cancel()
 conn,fingerprint,err:=hostclient.Connect(ctx,endpoint,"integration-session",key)
 if err!=nil{t.Fatalf("host handshake: %v",err)}
 defer conn.Close()
 expected,err:=identity.Fingerprint(key.Public)
 if err!=nil{t.Fatal(err)}
 if fingerprint!=expected{t.Fatalf("fingerprint mismatch: %q != %q",fingerprint,expected)}
 // Identity authentication alone must never authorize application traffic.
 if err:=conn.WriteMessage(websocket.TextMessage,[]byte(`{"type":"offer"}`));err!=nil{t.Fatal(err)}
 _=conn.SetReadDeadline(time.Now().Add(3*time.Second))
 _,_,err=conn.ReadMessage()
 if err==nil{t.Fatal("server accepted unauthorized application message")}
 var closeErr *websocket.CloseError
 if !errors.As(err,&closeErr)||closeErr.Code!=websocket.ClosePolicyViolation{
  t.Fatalf("expected policy violation close, got: %v",err)
 }
}
