package main

import (
 "encoding/json"
 "errors"
 "net"
 "net/http"
 "strings"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/session"
)

type pairingAPI struct {store *session.Store;revoke func(string,string)error}

func newPairingAPI() *pairingAPI {
 s,err:=session.NewStore(5*time.Minute)
 if err!=nil {panic(err)}
 return &pairingAPI{store:s}
}

func jsonReply(w http.ResponseWriter,status int,v any){
 w.Header().Set("Content-Type","application/json")
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.WriteHeader(status)
 _=json.NewEncoder(w).Encode(v)
}
func readID(r *http.Request) string {
 return r.PathValue("id")
}
func credential(r *http.Request)string{
 h:=r.Header.Get("Authorization")
 if !strings.HasPrefix(h,"Bearer "){return ""}
 return strings.TrimPrefix(h,"Bearer ")
}
func (a *pairingAPI) routes(mux *http.ServeMux){
 // Defense in depth: development API is inaccessible to non-loopback peers.
 onlyLocal:=func(next http.HandlerFunc)http.HandlerFunc{return func(w http.ResponseWriter,r *http.Request){
  host,_,err:=net.SplitHostPort(r.RemoteAddr)
  if err!=nil || !net.ParseIP(host).IsLoopback(){jsonReply(w,403,map[string]string{"error":"loopback only"});return}
  next(w,r)
 }}
 mux.HandleFunc("POST /v1/pairings",onlyLocal(func(w http.ResponseWriter,r *http.Request){
  id,host,err:=a.store.Create()
  if err!=nil{jsonReply(w,500,map[string]string{"error":"internal error"});return}
  jsonReply(w,201,map[string]string{"id":id,"host_capability":host})
 }))
 mux.HandleFunc("POST /v1/pairings/{id}/request",onlyLocal(func(w http.ResponseWriter,r *http.Request){
  tech,err:=a.store.Request(readID(r))
  if err!=nil{pairError(w,err);return}
  jsonReply(w,200,map[string]string{"technician_capability":tech,"state":"awaiting_approval"})
 }))
 mux.HandleFunc("POST /v1/pairings/{id}/approve",onlyLocal(func(w http.ResponseWriter,r *http.Request){
  if err:=a.store.Decide(readID(r),credential(r),true);err!=nil{pairError(w,err);return}
  jsonReply(w,200,map[string]string{"state":"approved"})
 }))
 mux.HandleFunc("POST /v1/pairings/{id}/deny",onlyLocal(func(w http.ResponseWriter,r *http.Request){
  if err:=a.store.Decide(readID(r),credential(r),false);err!=nil{pairError(w,err);return}
  jsonReply(w,200,map[string]string{"state":"denied"})
 }))
 mux.HandleFunc("POST /v1/pairings/{id}/revoke",onlyLocal(func(w http.ResponseWriter,r *http.Request){
  revoke:=a.store.Revoke
  if a.revoke!=nil{revoke=a.revoke}
  if err:=revoke(readID(r),credential(r));err!=nil{pairError(w,err);return}
  jsonReply(w,200,map[string]string{"state":"revoked"})
 }))
 mux.HandleFunc("GET /v1/pairings/{id}/authorization",onlyLocal(func(w http.ResponseWriter,r *http.Request){
  if !a.store.Authorized(readID(r),credential(r)){jsonReply(w,403,map[string]string{"authorized":"false"});return}
  jsonReply(w,200,map[string]bool{"authorized":true})
 }))
}
func pairError(w http.ResponseWriter,err error){
 switch {
 case errors.Is(err,session.ErrUnauthorized):jsonReply(w,403,map[string]string{"error":"forbidden"})
 case errors.Is(err,session.ErrNotFound),errors.Is(err,session.ErrExpired):jsonReply(w,404,map[string]string{"error":"not found"})
 case errors.Is(err,session.ErrInvalidState):jsonReply(w,409,map[string]string{"error":"invalid state"})
 default:jsonReply(w,500,map[string]string{"error":"internal error"})
 }
}
