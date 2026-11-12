package hostclient

import (
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestScreenShareStatusRequiresTokenAndReflectsStop(t *testing.T){
 b,err:=NewConsentBridge("http://127.0.0.1:9876","session",func(string,bool)error{return nil})
 if err!=nil{t.Fatal(err)}
 b.SetScreenShareStatus(func()string{return "waiting"})
 get:=func(token string)(int,string){
  t.Helper()
  r:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:9876/api/screen-share/status",nil)
  r.RemoteAddr="127.0.0.1:1234"
  r.Header.Set("X-PloosAssist-Token",token)
  w:=httptest.NewRecorder()
  b.UIHandler().ServeHTTP(w,r)
  var result struct{State string `json:"state"`}
  _=json.Unmarshal(w.Body.Bytes(),&result)
  return w.Code,result.State
 }
 if code,_:=get("wrong");code!=http.StatusForbidden{t.Fatalf("unauthorized: %d",code)}
 if code,state:=get(b.Token());code!=200||state!="waiting"{t.Fatalf("waiting: %d %s",code,state)}
 b.mu.Lock();b.stopped=true;b.mu.Unlock()
 if code,state:=get(b.Token());code!=200||state!="stopped"{t.Fatalf("stopped: %d %s",code,state)}
}
