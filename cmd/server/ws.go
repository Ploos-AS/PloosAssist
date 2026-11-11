package main

import (
 "crypto/ed25519"
 "encoding/base64"
 "encoding/json"
 "net/http"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/Ploos-AS/PloosAssist/internal/signaling"
 "github.com/Ploos-AS/PloosAssist/internal/wsproto"
)

// wsAPI is a one-shot, loopback-only proof-of-possession experiment.
// No signaling messages are accepted or forwarded.
type wsAPI struct{
 auth *identity.TransportAuth
 connections *signaling.Connections
}
func newWSAPI()*wsAPI{
 a,err:=identity.NewTransportAuth(30*time.Second,128);if err!=nil{panic(err)}
 connections,err:=signaling.NewConnections(128,time.Minute)
 if err!=nil{panic(err)}
 return &wsAPI{auth:a,connections:connections}
}
func (a *wsAPI)routes(mux *http.ServeMux){
 mux.HandleFunc("GET /v1/ws",func(w http.ResponseWriter,r *http.Request){
  // Client key and session are provided only to set up a proof challenge;
  // neither field establishes trust or session membership.
  key,err:=base64.RawURLEncoding.DecodeString(r.URL.Query().Get("public_key"))
  sessionID:=r.URL.Query().Get("session_id")
  if err!=nil||len(key)!=ed25519.PublicKeySize||len(sessionID)==0||len(sessionID)>128{
   jsonReply(w,400,map[string]string{"error":"invalid identity parameters"});return
  }
  conn,rw,err:=wsproto.Upgrade(w,r)
  if err!=nil{jsonReply(w,400,map[string]string{"error":"invalid websocket upgrade"});return}
  defer conn.Close()
  _=conn.SetDeadline(time.Now().Add(15*time.Second))
  id,err:=a.auth.Open(sessionID,ed25519.PublicKey(key));if err!=nil{return}
  defer a.auth.Close(id)
 defer a.connections.Close(id)
  challengeID,ch,err:=a.auth.Challenge(id);if err!=nil{return}
  start,_:=json.Marshal(map[string]string{"type":"challenge","nonce":base64.RawURLEncoding.EncodeToString(ch.Nonce),"challenge_id":challengeID})
  if err:=wsproto.WriteText(rw,start);err!=nil{return}
  if err:=rw.Flush();err!=nil{return}
  payload,err:=wsproto.ReadMessage(rw,rw);if err!=nil{return}
  var proof struct{Type string `json:"type"`;Signature string `json:"signature"`}
  if json.Unmarshal(payload,&proof)!=nil||proof.Type!="proof"{return}
  sig,err:=base64.RawURLEncoding.DecodeString(proof.Signature);if err!=nil{return}
  fingerprint,err:=a.auth.Authenticate(id,challengeID,sig)
  if err!=nil{return}
  verifiedSession,verifiedFingerprint,err:=a.auth.Identity(id)
  if err!=nil || verifiedSession!=sessionID || verifiedFingerprint!=fingerprint{return}
  if err:=a.connections.Register(id,verifiedSession,verifiedFingerprint);err!=nil{return}
  response,_:=json.Marshal(map[string]string{"type":"proof_verified","fingerprint":fingerprint})
  if wsproto.WriteText(rw,response)==nil{_ = rw.Flush()}
  // Disconnect immediately: no signaling, media, or remote input is enabled.
 })
}
