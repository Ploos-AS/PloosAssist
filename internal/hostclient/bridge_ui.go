package hostclient

import (
 "embed"
 "net"
 "net/http"
 "strings"
)

//go:embed ui/index.html ui/style.css ui/script.js
var consentUI embed.FS

// UIHandler serves only bundled, local, same-origin assets and the
// capability-guarded consent API. Never serve this over a public interface.
func (b *ConsentBridge)UIHandler() http.Handler {
 api:=b.Handler()
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  w.Header().Set("Referrer-Policy","no-referrer")
  w.Header().Set("X-Frame-Options","DENY")
  w.Header().Set("Content-Security-Policy","default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
  host,_,err:=net.SplitHostPort(r.RemoteAddr)
  if err!=nil||net.ParseIP(host)==nil||!net.ParseIP(host).IsLoopback()||r.Host!=strings.TrimPrefix(b.origin,"http://"){
   http.Error(w,"forbidden",http.StatusForbidden);return
  }
  if strings.HasPrefix(r.URL.Path,"/api/"){api.ServeHTTP(w,r);return}
  if r.Method!="GET"||r.URL.RawQuery!=""{http.NotFound(w,r);return}
  switch r.URL.Path{
  case "/":
   data,err:=consentUI.ReadFile("ui/index.html")
   if err!=nil{http.Error(w,"asset unavailable",500);return}
   w.Header().Set("Content-Type","text/html; charset=utf-8")
   _,_=w.Write([]byte(strings.ReplaceAll(string(data),"__PLOOS_TOKEN__",b.token)))
  case "/style.css","/script.js":
   data,err:=consentUI.ReadFile("ui"+r.URL.Path)
   if err!=nil{http.NotFound(w,r);return}
   if r.URL.Path=="/style.css"{w.Header().Set("Content-Type","text/css; charset=utf-8")}else{w.Header().Set("Content-Type","application/javascript; charset=utf-8")}
   _,_=w.Write(data)
  default:http.NotFound(w,r)
  }
 })
}
