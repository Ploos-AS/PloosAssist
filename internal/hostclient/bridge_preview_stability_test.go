package hostclient

import (
 "errors"
 "net/http"
 "net/http/httptest"
 "sync"
 "testing"
)

func TestPreviewStopDuringCallbackFailsClosed(t *testing.T){
 b,err:=NewConsentBridge("http://127.0.0.1:7655","test-session",func(string,bool)error{return nil})
 if err!=nil{t.Fatal(err)}
 entered:=make(chan struct{})
 release:=make(chan struct{})
 b.SetScreenSharePreview(func()([]byte,error){
  close(entered)
  <-release
  return []byte{137,80,78,71},nil
 })
 req:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:7655/api/screen-share/preview.png",nil)
 req.RemoteAddr="127.0.0.1:34567"
 req.Host="127.0.0.1:7655"
 req.Header.Set("X-PloosAssist-Token",b.Token())
 var wg sync.WaitGroup
 var got int
 wg.Add(1)
 go func(){
  defer wg.Done()
  w:=httptest.NewRecorder()
  b.Handler().ServeHTTP(w,req)
  got=w.Code
 }()
 <-entered
 b.mu.Lock();b.stopped=true;b.mu.Unlock()
 close(release)
 wg.Wait()
 if got!=http.StatusGone{t.Fatalf("preview escaped after stop: %d",got)}
}
func TestPreviewCallbackFailureDoesNotLeakFrame(t *testing.T){
 b,err:=NewConsentBridge("http://127.0.0.1:7656","test-session",func(string,bool)error{return nil})
 if err!=nil{t.Fatal(err)}
 b.SetScreenSharePreview(func()([]byte,error){return nil,errors.New("no frame")})
 req:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:7656/api/screen-share/preview.png",nil)
 req.RemoteAddr="127.0.0.1:34567"
 req.Host="127.0.0.1:7656"
 req.Header.Set("X-PloosAssist-Token",b.Token())
 w:=httptest.NewRecorder()
 b.Handler().ServeHTTP(w,req)
 if w.Code!=http.StatusServiceUnavailable{t.Fatalf("callback failure: %d",w.Code)}
}
