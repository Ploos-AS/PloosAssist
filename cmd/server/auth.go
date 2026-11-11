package main

import (
 "crypto/ed25519"
 "encoding/base64"
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/identity"
)

type authAPI struct {state *identity.TransportAuth}

func newAuthAPI()*authAPI{
 state,err:=identity.NewTransportAuth(30*time.Second,128)
 if err!=nil{panic(err)}
 return &authAPI{state:state}
}

func decodeAuthJSON(w http.ResponseWriter,r *http.Request,v any)bool{
 if r.Header.Get("Content-Type")!="application/json" {
  jsonReply(w,415,map[string]string{"error":"application/json required"});return false
 }
 d:=json.NewDecoder(io.LimitReader(r.Body,4097))
 d.DisallowUnknownFields()
 if err:=d.Decode(v);err!=nil{jsonReply(w,400,map[string]string{"error":"invalid JSON"});return false}
 var extra any
 if err:=d.Decode(&extra);err!=io.EOF{jsonReply(w,400,map[string]string{"error":"trailing or oversized JSON"});return false}
 return true
}

func (a *authAPI)routes(mux *http.ServeMux){
 mux.HandleFunc("POST /v1/auth/start",func(w http.ResponseWriter,r *http.Request){
  var req struct {SessionID string `json:"session_id"`;PublicKey string `json:"public_key"`}
  if !decodeAuthJSON(w,r,&req){return}
  key,err:=base64.RawURLEncoding.DecodeString(req.PublicKey)
  if err!=nil || len(key)!=ed25519.PublicKeySize || len(req.SessionID)==0 || len(req.SessionID)>128{
   jsonReply(w,400,map[string]string{"error":"invalid session or public key"});return
  }
  connectionID,err:=a.state.Open(req.SessionID,ed25519.PublicKey(key))
  if err!=nil{jsonReply(w,500,map[string]string{"error":"could not create connection"});return}
  challengeID,challenge,err:=a.state.Challenge(connectionID)
  if err!=nil{a.state.Close(connectionID);jsonReply(w,503,map[string]string{"error":"challenge unavailable"});return}
  jsonReply(w,201,map[string]any{"connection_id":connectionID,"challenge_id":challengeID,"nonce":base64.RawURLEncoding.EncodeToString(challenge.Nonce),"expires_at":challenge.ExpiresAt.UTC().Format(time.RFC3339Nano)})
 })
 mux.HandleFunc("POST /v1/auth/verify",func(w http.ResponseWriter,r *http.Request){
  var req struct {ConnectionID string `json:"connection_id"`;ChallengeID string `json:"challenge_id"`;Signature string `json:"signature"`}
  if !decodeAuthJSON(w,r,&req){return}
  signature,err:=base64.RawURLEncoding.DecodeString(req.Signature)
  if err!=nil || len(signature)!=ed25519.SignatureSize{
   jsonReply(w,400,map[string]string{"error":"invalid signature encoding"});return
  }
  fingerprint,err:=a.state.Authenticate(req.ConnectionID,req.ChallengeID,signature)
  if err!=nil{
   if errors.Is(err,identity.ErrConnectionUnknown) {jsonReply(w,404,map[string]string{"error":"unknown connection"});return}
   jsonReply(w,403,map[string]string{"error":"proof rejected"});return
  }
  // This response is informational: it is NOT a bearer token or authorization to signal.
  jsonReply(w,200,map[string]string{"fingerprint":fingerprint,"status":"proof_verified"})
 })
}
