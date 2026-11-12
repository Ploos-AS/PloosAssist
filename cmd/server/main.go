// PloosAssist signaling server foundation.
// This intentionally does not accept sessions or remote control commands.
package main

import (
 "encoding/json"
 "flag"
 "log"
 "net/http"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/signaling"
)

func handler() http.Handler {
 mux:=http.NewServeMux()
 mux.HandleFunc("GET /healthz",func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
 })
 pairing:=newPairingAPI()
 ws:=newWSAPI()
 ws.enrollment=mustEnrollment(pairing.store,ws.connections,ws.roles)
 var err error
 ws.revocation,err=signaling.NewRevocation(pairing.store,ws.consent,ws.roles,ws.approvals,ws.events,ws.delivery,ws.router)
 if err!=nil{panic(err)}
 pairing.revoke=func(sessionID,capability string)error{
  if err:=ws.revocation.Revoke(sessionID,capability);err!=nil{return err}
  ws.terminateSession(sessionID)
  return nil
 }
 pairing.routes(mux)
 newAuthAPI().routes(mux)
 ws.routes(mux)
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
