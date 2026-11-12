package hostclient

import (
 "crypto/rand"
 "encoding/base64"
 "encoding/json"
 "errors"
 "net"
 "net/http"
 "strings"
 "sync"
 "time"
)

// ConsentBridge is a loopback-only, same-origin HTTP boundary. It never
// accepts arbitrary ticket IDs from the browser as authorization.
type ConsentBridge struct {
 mu sync.Mutex
 origin string
 token string
 session string
 pending *PendingConsent
 decide func(string,bool)error
 stop func()error
 screenShare func()error
 screenShareStatus func()string
 screenSharePreview func()([]byte,error)
 stopped bool
}
func NewConsentBridge(origin,session string,decide func(string,bool)error)(*ConsentBridge,error){
 if !strings.HasPrefix(origin,"http://127.0.0.1:")||session==""||decide==nil{return nil,errors.New("invalid local consent bridge")}
 raw:=make([]byte,32)
 if _,err:=rand.Read(raw);err!=nil{return nil,err}
 return &ConsentBridge{origin:origin,token:base64.RawURLEncoding.EncodeToString(raw),session:session,decide:decide},nil
}
// SetPending must be called only by the authenticated host WebSocket owner.
func (b *ConsentBridge)SetPending(p PendingConsent){
 b.mu.Lock();defer b.mu.Unlock()
 b.pending=&p
}
func (b *ConsentBridge)Clear(){
 b.mu.Lock();defer b.mu.Unlock()
 b.pending=nil
}
// SetStop installs the host-owned session termination callback.
func (b *ConsentBridge)SetStop(stop func()error){b.mu.Lock();defer b.mu.Unlock();b.stop=stop}
// Handler is intentionally API-only. A future trusted host process must serve
// the HTML/JS from the SAME origin and provision the token without URL params.
func (b *ConsentBridge)Handler()http.Handler{
 mux:=http.NewServeMux()
 guard:=func(w http.ResponseWriter,r *http.Request)bool{
  w.Header().Set("Cache-Control","no-store")
  w.Header().Set("X-Content-Type-Options","nosniff")
  w.Header().Set("Access-Control-Allow-Origin","")
  host,_,err:=net.SplitHostPort(r.RemoteAddr)
  if err!=nil||net.ParseIP(host)==nil||!net.ParseIP(host).IsLoopback()||r.Host!=strings.TrimPrefix(b.origin,"http://")||(r.Header.Get("Origin")!=b.origin&&(r.Method!="GET"||r.Header.Get("Origin")!=""))||
   r.Header.Get("X-PloosAssist-Token")!=b.token{
   http.Error(w,"forbidden",http.StatusForbidden);return false
  }
  return true
 }
 mux.HandleFunc("POST /api/consent/stop",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  b.mu.Lock()
  defer b.mu.Unlock()
  if b.stop==nil{http.Error(w,"unavailable",503);return}
  // Fail closed locally even if the remote server is unreachable.
  b.pending=nil
  b.stopped=true
  if err:=b.stop();err!=nil{http.Error(w,"stop failed",502);return}
  _=json.NewEncoder(w).Encode(map[string]string{"type":"session_stopped"})
 })
 mux.HandleFunc("GET /api/consent",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  b.mu.Lock();defer b.mu.Unlock()
  if b.pending==nil{_ =json.NewEncoder(w).Encode(map[string]string{"type":"approval_none"});return}
  expires,err:=time.Parse(time.RFC3339,b.pending.ExpiresAt)
  if err!=nil||!time.Now().Before(expires){
   b.pending=nil
   _=json.NewEncoder(w).Encode(map[string]string{"type":"approval_none"});return
  }
  _=json.NewEncoder(w).Encode(map[string]string{
   "type":"approval_pending","sessionId":b.session,
   "ticketId":b.pending.TicketID,"technicianFingerprint":b.pending.TechnicianFingerprint,
   "expiresAt":b.pending.ExpiresAt,
  })
 })
 mux.HandleFunc("POST /api/consent/decision",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  r.Body=http.MaxBytesReader(w,r.Body,512)
  var request struct{TicketID string `json:"ticketId"`;Approve *bool `json:"approve"`}
  decoder:=json.NewDecoder(r.Body)
  decoder.DisallowUnknownFields()
  if decoder.Decode(&request)!=nil||request.Approve==nil{http.Error(w,"bad request",400);return}
  b.mu.Lock()
  defer b.mu.Unlock()
  if b.pending==nil||request.TicketID!=b.pending.TicketID{
   http.Error(w,"no pending request",409);return
  }
  expiry,err:=time.Parse(time.RFC3339,b.pending.ExpiresAt)
  if err!=nil||!time.Now().Before(expiry){
   b.pending=nil
   http.Error(w,"expired",409);return
  }
  // Hold lock while deciding so the same ticket cannot race twice.
  pending:=b.pending
  b.pending=nil
  if err:=b.decide(pending.TicketID,*request.Approve);err!=nil{
   http.Error(w,"decision failed",502);return
  }
  _=json.NewEncoder(w).Encode(map[string]string{"type":"decision_recorded"})
 })
 b.screenShareRoutes(mux,guard)
 return mux
}
// Token is a secret bootstrap credential for a trusted local HTML renderer.
// Do not log it, put it in a URL, or expose it cross-origin.
func (b *ConsentBridge)Token()string{return b.token}
