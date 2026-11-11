package main

import (
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
)

func TestGuardLimitAndReset(t *testing.T){
 g:=newRequestGuard(2)
 clock:=time.Date(2026,1,1,0,0,0,0,time.UTC)
 g.now=func()time.Time{return clock}
 if !g.allow()||!g.allow()||g.allow(){t.Fatal("limit not enforced")}
 clock=clock.Add(time.Minute)
 if !g.allow(){t.Fatal("window did not reset")}
}
func TestGuardRejectsBrowserOrigin(t *testing.T){
 h:=newRequestGuard(20).wrap(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(204)}))
 r:=httptest.NewRequest("POST","/v1/pairings",nil)
 r.RemoteAddr="127.0.0.1:12345"
 r.Header.Set("Origin","https://example.org")
 w:=httptest.NewRecorder();h.ServeHTTP(w,r)
 if w.Code!=403{t.Fatalf("got %d",w.Code)}
}
func TestListenAddress(t *testing.T){
 cases:=map[string]bool{
  "127.0.0.1:8080":true,
  "[::1]:8080":true,
  "0.0.0.0:8080":false,
  ":8080":false,
  "192.0.2.1:8080":false,
  "localhost:8080":false,
 }
 for addr,want:=range cases{
  if got:=isLoopbackListenAddr(addr);got!=want{t.Errorf("%s: got %v, want %v",addr,got,want)}
 }
}
