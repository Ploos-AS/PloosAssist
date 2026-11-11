// PloosAssist signaling server foundation.
// This intentionally does not accept sessions or remote control commands.
package main

import (
 "encoding/json"
 "flag"
 "log"
 "net/http"
 "time"
)

func handler() http.Handler {
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /healthz",func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
 })
 newPairingAPI().routes(mux)
 return newRequestGuard(120).wrap(mux)
}

func main(){
 addr:=flag.String("listen","127.0.0.1:8080","listen address (loopback by default)")
 flag.Parse()
 if !isLoopbackListenAddr(*addr) {log.Fatal("development server requires numeric loopback listen address")}
 srv:=&http.Server{Addr:*addr,Handler:handler(),ReadHeaderTimeout:5*time.Second}
 log.Printf("PloosAssist development server listening on %s",*addr)
 log.Fatal(srv.ListenAndServe())
}
