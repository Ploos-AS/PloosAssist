package hostclient

import (
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestPreviewEndpointRequiresTokenAndStops(t *testing.T){
 b,err:=NewConsentBridge("http://127.0.0.1:7654","test-session",func(string,bool)error{return nil})
 if err!=nil{t.Fatal(err)}
 b.SetScreenSharePreview(func()([]byte,error){return []byte{137,80,78,71},nil})
 invoke:=func(token string)*httptest.ResponseRecorder{
  r:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:7654/api/screen-share/preview.png",nil)
  r.RemoteAddr="127.0.0.1:43210"
  r.Host="127.0.0.1:7654"
  if token!=""{r.Header.Set("X-PloosAssist-Token",token)}
  w:=httptest.NewRecorder()
  b.Handler().ServeHTTP(w,r)
  return w
 }
 if got:=invoke("").Code;got!=http.StatusForbidden{t.Fatalf("unauthorized: %d",got)}
 good:=invoke(b.Token())
 if good.Code!=http.StatusOK||good.Header().Get("Content-Type")!="image/png"||
  good.Header().Get("Cache-Control")!="no-store"{t.Fatalf("valid preview: %d %#v",good.Code,good.Header())}
 b.mu.Lock();b.stopped=true;b.mu.Unlock()
 if got:=invoke(b.Token()).Code;got!=http.StatusGone{t.Fatalf("after stop: %d",got)}
}
