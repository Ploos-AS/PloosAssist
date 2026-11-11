package identity

import (
 "errors"
 "testing"
 "time"
)

func TestTransportAuthentication(t *testing.T){
 a,err:=NewTransportAuth(time.Minute,10);if err!=nil{t.Fatal(err)}
 kp,_:=Generate()
 id,err:=a.Open("session",kp.Public);if err!=nil{t.Fatal(err)}
 if _,_,err:=a.Identity(id);!errors.Is(err,ErrConnectionUnauthenticated){t.Fatalf("unauthenticated accepted: %v",err)}
 challengeID,ch,err:=a.Challenge(id);if err!=nil{t.Fatal(err)}
 sig,_:=Sign(kp.Private,ch)
 fp,err:=a.Authenticate(id,challengeID,sig);if err!=nil{t.Fatal(err)}
 session,identity,err:=a.Identity(id);if err!=nil||session!="session"||identity!=fp{t.Fatalf("identity mismatch: %s %s %v",session,identity,err)}
 if _,err:=a.Authenticate(id,challengeID,sig);!errors.Is(err,ErrConnectionAlreadyAuthenticated){t.Fatalf("re-auth accepted: %v",err)}
 a.Close(id)
 if _,_,err:=a.Identity(id);!errors.Is(err,ErrConnectionUnknown){t.Fatalf("closed connection accepted: %v",err)}
}
func TestTransportRejectsCrossConnectionProof(t *testing.T){
 a,_:=NewTransportAuth(time.Minute,10)
 kp,_:=Generate()
 id1,_:=a.Open("s1",kp.Public)
 id2,_:=a.Open("s2",kp.Public)
 cid,ch,_:=a.Challenge(id1)
 sig,_:=Sign(kp.Private,ch)
 if _,err:=a.Authenticate(id2,cid,sig);!errors.Is(err,ErrChallengeContext){t.Fatalf("cross-connection accepted: %v",err)}
 if _,err:=a.Authenticate(id1,cid,sig);!errors.Is(err,ErrChallengeNotFound){t.Fatalf("reused challenge accepted: %v",err)}
}
func TestTransportRejectsWrongKey(t *testing.T){
 a,_:=NewTransportAuth(time.Minute,10)
 kp,_:=Generate();other,_:=Generate()
 id,_:=a.Open("s",kp.Public)
 cid,ch,_:=a.Challenge(id)
 sig,_:=Sign(other.Private,ch)
 if _,err:=a.Authenticate(id,cid,sig);!errors.Is(err,ErrInvalidSignature){t.Fatalf("wrong key accepted: %v",err)}
 if _,_,err:=a.Identity(id);!errors.Is(err,ErrConnectionUnauthenticated){t.Fatalf("failed proof authenticated: %v",err)}
}
