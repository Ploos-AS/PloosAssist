package hostclient

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)
func TestUIHandlerSameOriginAssetsAndTokenProtection(t *testing.T){
 bridge,err:=NewConsentBridge("http://127.0.0.1:9876","demo",func(string,bool)error{return nil})
 if err!=nil{t.Fatal(err)}
 send:=func(path,host string)*httptest.ResponseRecorder{
  t.Helper()
  req:=httptest.NewRequest(http.MethodGet,"http://127.0.0.1:9876"+path,nil)
  req.Host=host
  req.RemoteAddr="127.0.0.1:32145"
  w:=httptest.NewRecorder()
  bridge.UIHandler().ServeHTTP(w,req)
  return w
 }
 page:=send("/","127.0.0.1:9876")
 if page.Code!=200||!strings.Contains(page.Body.String(),bridge.Token()){t.Fatal("local token bootstrap unavailable")}
 if strings.Contains(page.Body.String(),"__PLOOS_TOKEN__"){t.Fatal("token placeholder not replaced")}
 if page.Header().Get("Content-Security-Policy")==""{t.Fatal("missing CSP")}
 if send("/","evil.example").Code!=403{t.Fatal("DNS rebinding host accepted")}
 if send("/script.js","127.0.0.1:9876").Code!=200{t.Fatal("script not served")}
 if send("/style.css","127.0.0.1:9876").Code!=200{t.Fatal("styles not served")}
 if send("/api/consent","127.0.0.1:9876").Code!=403{t.Fatal("API accessible without token")}
}
