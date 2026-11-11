package main

import (
 "net"
 "net/http"
 "strings"
 "sync"
 "time"
)

// requestGuard is a coarse, process-local safety limit for the development API.
// It is NOT a replacement for identity-aware abuse prevention.
type requestGuard struct {
 mu sync.Mutex
 window time.Time
 count int
 limit int
 now func() time.Time
}

func newRequestGuard(limit int) *requestGuard {
 return &requestGuard{limit:limit,now:time.Now}
}

func (g *requestGuard) allow() bool {
 g.mu.Lock()
 defer g.mu.Unlock()
 now:=g.now()
 if g.window.IsZero() || !now.Before(g.window.Add(time.Minute)) {
  g.window=now
  g.count=0
 }
 if g.count>=g.limit{return false}
 g.count++
 return true
}

func (g *requestGuard) wrap(next http.Handler)http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  w.Header().Set("Referrer-Policy","no-referrer")
  if !strings.HasPrefix(r.URL.Path,"/v1/pairings") {
   next.ServeHTTP(w,r)
   return
  }
  // Browser-based cross-origin requests are not supported by the dev API.
  if r.Header.Get("Origin")!="" {
   jsonReply(w,http.StatusForbidden,map[string]string{"error":"browser origins not supported"})
   return
  }
  host,_,err:=net.SplitHostPort(r.RemoteAddr)
  ip:=net.ParseIP(host)
  if err!=nil || ip==nil || !ip.IsLoopback(){
   jsonReply(w,http.StatusForbidden,map[string]string{"error":"loopback only"})
   return
  }
  if !g.allow(){
   w.Header().Set("Retry-After","60")
   jsonReply(w,http.StatusTooManyRequests,map[string]string{"error":"rate limit exceeded"})
   return
  }
  next.ServeHTTP(w,r)
 })
}

// isLoopbackListenAddr rejects unspecified, wildcard, and non-loopback binds.
// This does not protect against reverse proxies or forwarded connections.
func isLoopbackListenAddr(addr string)bool{
 host,_,err:=net.SplitHostPort(addr)
 if err!=nil{return false}
 ip:=net.ParseIP(host)
 return ip!=nil && ip.IsLoopback()
}
