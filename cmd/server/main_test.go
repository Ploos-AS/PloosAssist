package main

import (
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestHealth(t *testing.T) {
 r:=httptest.NewRequest(http.MethodGet,"/healthz",nil)
 w:=httptest.NewRecorder()
 handler().ServeHTTP(w,r)
 if w.Code!=http.StatusOK {t.Fatalf("got status %d",w.Code)}
 if w.Body.String()!="{\"status\":\"ok\"}\n" {t.Fatalf("unexpected body %q",w.Body.String())}
}
