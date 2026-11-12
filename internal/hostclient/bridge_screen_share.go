package hostclient

import (
 "encoding/json"
 "net/http"
)

// SetScreenShare installs a trusted host-owned action that starts the
// desktop portal prompt. It does not grant any media transmission.
func(b *ConsentBridge)SetScreenShare(start func()error){
 b.mu.Lock()
 defer b.mu.Unlock()
 b.screenShare=start
}
func(b *ConsentBridge)screenShareRoutes(mux *http.ServeMux,guard func(http.ResponseWriter,*http.Request)bool){
 mux.HandleFunc("POST /api/screen-share/start",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  b.mu.Lock()
  callback:=b.screenShare
  stopped:=b.stopped
  b.mu.Unlock()
  if stopped{http.Error(w,"session stopped",http.StatusConflict);return}
  if callback==nil{http.Error(w,"unavailable",http.StatusServiceUnavailable);return}
  if err:=callback();err!=nil{http.Error(w,"screen share unavailable",http.StatusConflict);return}
  _=json.NewEncoder(w).Encode(map[string]string{"type":"screen_share_requested"})
 })
}
