package hostclient

import (
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestScreenShareRequiresTokenOriginAndExplicitClick(t *testing.T){
 b,err:=NewConsentBridge("http://127.0.0.1:9876","session",func(string,bool)error{return nil})
 if err!=nil{t.Fatal(err)}
 calls:=0
 b.SetScreenShare(func()error{calls++;return nil})
 send:=func(origin,token string)int{
  t.Helper()
  r:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:9876/api/screen-share/start",nil)
  r.RemoteAddr="127.0.0.1:1234"
  r.Header.Set("Origin",origin)
  r.Header.Set("X-PloosAssist-Token",token)
  w:=httptest.NewRecorder()
  b.UIHandler().ServeHTTP(w,r)
  return w.Code
 }
 if got:=send("http://evil.invalid",b.Token());got!=403{t.Fatalf("origin: %d",got)}
 if got:=send("http://127.0.0.1:9876","wrong");got!=403{t.Fatalf("token: %d",got)}
 if calls!=0{t.Fatal("unauthorized portal callback")}
 if got:=send("http://127.0.0.1:9876",b.Token());got!=200{t.Fatalf("authorized request: %d",got)}
 if calls!=1{t.Fatalf("portal callback count %d",calls)}
 b.mu.Lock();b.stopped=true;b.mu.Unlock()
 if got:=send("http://127.0.0.1:9876",b.Token());got!=409{t.Fatalf("request after stop: %d",got)}
 if calls!=1{t.Fatal("portal restarted after stop")}
}
