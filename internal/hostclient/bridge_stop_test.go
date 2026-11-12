package hostclient

import (
 "net/http"
 "net/http/httptest"
 "testing"
)
func TestConsentBridgeStopRequiresTokenAndOrigin(t *testing.T){
 calls:=0
 b,err:=NewConsentBridge("http://127.0.0.1:9876","demo",func(string,bool)error{return nil})
 if err!=nil{t.Fatal(err)}
 b.SetStop(func()error{calls++;return nil})
 send:=func(origin,token string)int{
  t.Helper()
  r:=httptest.NewRequest(http.MethodPost,"http://127.0.0.1:9876/api/consent/stop",nil)
  r.RemoteAddr="127.0.0.1:32145"
  r.Header.Set("Origin",origin)
  r.Header.Set("X-PloosAssist-Token",token)
  w:=httptest.NewRecorder()
  b.UIHandler().ServeHTTP(w,r)
  return w.Code
 }
 if code:=send("http://attacker.example",b.Token());code!=403{t.Fatalf("wrong origin: %d",code)}
 if code:=send("http://127.0.0.1:9876","invalid");code!=403{t.Fatalf("wrong token: %d",code)}
 if calls!=0{t.Fatal("unauthorized stop callback")}
 if code:=send("http://127.0.0.1:9876",b.Token());code!=200{t.Fatalf("authorized stop: %d",code)}
 if calls!=1{t.Fatalf("expected one stop callback, got %d",calls)}
}
