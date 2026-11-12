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
func(b *ConsentBridge)SetScreenShareStatus(status func()string){
 b.mu.Lock();defer b.mu.Unlock();b.screenShareStatus=status
}
// SetScreenSharePreview installs a trusted local-only PNG snapshot callback.
func(b *ConsentBridge)SetScreenSharePreview(preview func()([]byte,error)){
 b.mu.Lock();defer b.mu.Unlock();b.screenSharePreview=preview
}
func(b *ConsentBridge)screenShareRoutes(mux *http.ServeMux,guard func(http.ResponseWriter,*http.Request)bool){
 mux.HandleFunc("GET /api/screen-share/status",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  b.mu.Lock();callback:=b.screenShareStatus;stopped:=b.stopped;b.mu.Unlock()
  state:="unavailable"
  if stopped{state="stopped"}else if callback!=nil{state=callback()}
  _=json.NewEncoder(w).Encode(map[string]string{"type":"screen_share_status","state":state})
 })
 mux.HandleFunc("GET /api/screen-share/preview.png",func(w http.ResponseWriter,r *http.Request){
  if !guard(w,r){return}
  b.mu.Lock();callback:=b.screenSharePreview;stopped:=b.stopped;b.mu.Unlock()
  if stopped{http.Error(w,"session stopped",http.StatusGone);return}
  if callback==nil{http.Error(w,"preview unavailable",http.StatusServiceUnavailable);return}
  png,err:=callback()
  if err!=nil||len(png)==0{http.Error(w,"no local frame",http.StatusServiceUnavailable);return}
  b.mu.Lock();stopped=b.stopped;b.mu.Unlock()
  if stopped{http.Error(w,"session stopped",http.StatusGone);return}
  w.Header().Set("Content-Type","image/png")
  w.Header().Set("Content-Security-Policy","default-src 'none'; sandbox")
  _,_=w.Write(png)
 })
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
