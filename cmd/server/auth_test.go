package main

import (
 "bytes"
 "crypto/ed25519"
 "encoding/base64"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/Ploos-AS/PloosAssist/internal/identity"
)

func authCall(t *testing.T,h http.Handler,path string,body any)(int,map[string]string){
 t.Helper()
 data,err:=json.Marshal(body);if err!=nil{t.Fatal(err)}
 r:=httptest.NewRequest("POST",path,bytes.NewReader(data))
 r.RemoteAddr="127.0.0.1:30000"
 r.Header.Set("Content-Type","application/json")
 w:=httptest.NewRecorder();h.ServeHTTP(w,r)
 var result map[string]string
 if err:=json.Unmarshal(w.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
 return w.Code,result
}
func TestAuthHandshake(t *testing.T){
 h:=handler()
 kp,err:=identity.Generate();if err!=nil{t.Fatal(err)}
 status,start:=authCall(t,h,"/v1/auth/start",map[string]string{"session_id":"dev-session","public_key":base64.RawURLEncoding.EncodeToString(kp.Public)})
 if status!=201{t.Fatalf("start: %d %v",status,start)}
 nonce,err:=base64.RawURLEncoding.DecodeString(start["nonce"]);if err!=nil{t.Fatal(err)}
 challenge:=identity.Challenge{Nonce:nonce}
 sig,err:=identity.Sign(kp.Private,challenge);if err!=nil{t.Fatal(err)}
 proof:=map[string]string{"connection_id":start["connection_id"],"challenge_id":start["challenge_id"],"signature":base64.RawURLEncoding.EncodeToString(sig)}
 status,result:=authCall(t,h,"/v1/auth/verify",proof)
 if status!=200||result["status"]!="proof_verified"{t.Fatalf("verify: %d %v",status,result)}
 status,_=authCall(t,h,"/v1/auth/verify",proof)
 if status==200{t.Fatal("replay accepted")}
}
func TestAuthWrongKey(t *testing.T){
 h:=handler()
 kp,_:=identity.Generate();other,_:=identity.Generate()
 _,start:=authCall(t,h,"/v1/auth/start",map[string]string{"session_id":"dev-session","public_key":base64.RawURLEncoding.EncodeToString(kp.Public)})
 nonce,_:=base64.RawURLEncoding.DecodeString(start["nonce"])
 sig:=ed25519.Sign(other.Private,append([]byte("PloosAssist/identity/v1\x00"),nonce...))
 status,_:=authCall(t,h,"/v1/auth/verify",map[string]string{"connection_id":start["connection_id"],"challenge_id":start["challenge_id"],"signature":base64.RawURLEncoding.EncodeToString(sig)})
 if status!=403{t.Fatalf("wrong key returned %d",status)}
}
func TestAuthRejectsNonLoopback(t *testing.T){
 h:=handler()
 r:=httptest.NewRequest("POST","/v1/auth/start",bytes.NewReader([]byte("{}")))
 r.RemoteAddr="192.0.2.5:9000"
 r.Header.Set("Content-Type","application/json")
 w:=httptest.NewRecorder();h.ServeHTTP(w,r)
 if w.Code!=403{t.Fatalf("non-loopback returned %d",w.Code)}
}
