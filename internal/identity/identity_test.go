package identity

import (
 "bytes"
 "errors"
 "testing"
 "time"
)

func TestChallengeProof(t *testing.T){
 now:=time.Date(2026,11,11,0,0,0,0,time.UTC)
 a,err:=Generate();if err!=nil{t.Fatal(err)}
 b,err:=Generate();if err!=nil{t.Fatal(err)}
 challenge,err:=NewChallenge(now,30*time.Second);if err!=nil{t.Fatal(err)}
 signature,err:=Sign(a.Private,challenge);if err!=nil{t.Fatal(err)}
 if err:=Verify(a.Public,challenge,signature,now);err!=nil{t.Fatal(err)}
 if err:=Verify(b.Public,challenge,signature,now);!errors.Is(err,ErrInvalidSignature){t.Fatalf("wrong key: %v",err)}
 if err:=Verify(a.Public,challenge,signature,now.Add(30*time.Second));!errors.Is(err,ErrExpired){t.Fatalf("expiry: %v",err)}
 another,err:=NewChallenge(now,30*time.Second);if err!=nil{t.Fatal(err)}
 if bytes.Equal(challenge.Nonce,another.Nonce){t.Fatal("duplicate nonce")}
 if err:=Verify(a.Public,another,signature,now);!errors.Is(err,ErrInvalidSignature){t.Fatalf("replayed challenge: %v",err)}
 f1,_:=Fingerprint(a.Public);f2,_:=Fingerprint(a.Public)
 if f1!=f2 || f1==""{t.Fatal("unstable fingerprint")}
}
func TestChallengeTTL(t *testing.T){
 for _,ttl:=range []time.Duration{0,-time.Second,2*time.Minute}{
  if _,err:=NewChallenge(time.Now(),ttl);err==nil{t.Fatalf("accepted ttl %v",ttl)}
 }
}
