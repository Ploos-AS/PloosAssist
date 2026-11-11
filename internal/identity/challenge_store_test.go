package identity

import (
 "errors"
 "sync"
 "sync/atomic"
 "testing"
 "time"
)

func TestChallengeStoreOneTime(t *testing.T){
 kp,err:=Generate();if err!=nil{t.Fatal(err)}
 s,err:=NewChallengeStore(30*time.Second,5);if err!=nil{t.Fatal(err)}
 id,ch,err:=s.Issue(kp.Public,"conn1","session1");if err!=nil{t.Fatal(err)}
 sig,err:=Sign(kp.Private,ch);if err!=nil{t.Fatal(err)}
 if err:=s.Consume(id,kp.Public,"conn1","session1",sig);err!=nil{t.Fatal(err)}
 if err:=s.Consume(id,kp.Public,"conn1","session1",sig);!errors.Is(err,ErrChallengeNotFound){t.Fatalf("replay accepted: %v",err)}
}
func TestChallengeStoreWrongContextConsumes(t *testing.T){
 kp,_:=Generate()
 s,_:=NewChallengeStore(time.Minute,2)
 id,ch,_:=s.Issue(kp.Public,"a","s")
 sig,_:=Sign(kp.Private,ch)
 if err:=s.Consume(id,kp.Public,"b","s",sig);!errors.Is(err,ErrChallengeContext){t.Fatalf("context accepted: %v",err)}
 if err:=s.Consume(id,kp.Public,"a","s",sig);!errors.Is(err,ErrChallengeNotFound){t.Fatalf("failed attempt did not consume: %v",err)}
}
func TestChallengeStoreExpiryAndCapacity(t *testing.T){
 kp,_:=Generate()
 s,_:=NewChallengeStore(time.Second,1)
 clock:=time.Now()
 s.now=func()time.Time{return clock}
 id,ch,_:=s.Issue(kp.Public,"a","s")
 if _,_,err:=s.Issue(kp.Public,"b","s");!errors.Is(err,ErrChallengeCapacity){t.Fatalf("capacity: %v",err)}
 sig,_:=Sign(kp.Private,ch)
 clock=clock.Add(2*time.Second)
 if err:=s.Consume(id,kp.Public,"a","s",sig);!errors.Is(err,ErrExpired){t.Fatalf("expiry: %v",err)}
 if _,_,err:=s.Issue(kp.Public,"b","s");err!=nil{t.Fatal(err)}
}
func TestChallengeStoreConcurrentConsumption(t *testing.T){
 kp,_:=Generate()
 s,_:=NewChallengeStore(time.Minute,2)
 id,ch,_:=s.Issue(kp.Public,"a","s")
 sig,_:=Sign(kp.Private,ch)
 var successes atomic.Int32
 var wg sync.WaitGroup
 for i:=0;i<20;i++{wg.Add(1);go func(){defer wg.Done();if s.Consume(id,kp.Public,"a","s",sig)==nil{successes.Add(1)}}()}
 wg.Wait()
 if successes.Load()!=1{t.Fatalf("got %d successful consumptions",successes.Load())}
}
