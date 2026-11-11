package session

import (
 "errors"
 "testing"
 "time"
)

func TestConsentLifecycle(t *testing.T){
 s,_:=NewStore(time.Minute)
 id,host,err:=s.Create();if err!=nil{t.Fatal(err)}
 if s.Authorized(id,""){t.Fatal("unapproved session authorized")}
 tech,err:=s.Request(id);if err!=nil{t.Fatal(err)}
 if s.Authorized(id,tech){t.Fatal("session authorized before consent")}
 if err:=s.Decide(id,"wrong",true);!errors.Is(err,ErrUnauthorized){t.Fatalf("expected unauthorized, got %v",err)}
 if err:=s.Decide(id,host,true);err!=nil{t.Fatal(err)}
 if !s.Authorized(id,tech){t.Fatal("approved capability rejected")}
 if s.Authorized(id,"wrong"){t.Fatal("incorrect capability accepted")}
 if err:=s.Revoke(id,host);err!=nil{t.Fatal(err)}
 if s.Authorized(id,tech){t.Fatal("revoked session authorized")}
}
func TestDenialAndExpiry(t *testing.T){
 s,_:=NewStore(time.Minute)
 id,host,_:=s.Create()
 tech,_:=s.Request(id)
 if err:=s.Decide(id,host,false);err!=nil{t.Fatal(err)}
 if s.Authorized(id,tech){t.Fatal("denied session authorized")}
 if _,err:=s.Request(id);!errors.Is(err,ErrInvalidState){t.Fatalf("expected invalid state, got %v",err)}
 id2,_,_:=s.Create()
 s.now=func()time.Time{return time.Now().Add(2*time.Minute)}
 if _,err:=s.Request(id2);!errors.Is(err,ErrExpired){t.Fatalf("expected expired, got %v",err)}
}
func TestTTLBounds(t *testing.T){
 for _,ttl:=range []time.Duration{0,-time.Second,16*time.Minute}{
  if _,err:=NewStore(ttl);err==nil{t.Fatalf("accepted ttl %s",ttl)}
 }
}
