package signaling

import (
 "errors"
 "sync"
 "testing"
 "time"
)
func TestConnectionsCapacityAndClose(t *testing.T){
 c,_:=NewConnections(1,time.Minute)
 if err:=c.Register("a","s","fingerprint");err!=nil{t.Fatal(err)}
 if err:=c.Register("b","s","fingerprint");!errors.Is(err,ErrConnectionLimit){t.Fatalf("capacity: %v",err)}
 session,fp,err:=c.Identity("a")
 if err!=nil||session!="s"||fp!="fingerprint"{t.Fatalf("identity: %q %q %v",session,fp,err)}
 c.Close("a")
 if c.Count()!=0{t.Fatal("connection leaked")}
 if err:=c.Register("b","s","fingerprint");err!=nil{t.Fatal(err)}
}
func TestConnectionsExpiry(t *testing.T){
 c,_:=NewConnections(1,time.Minute)
 now:=time.Now()
 c.now=func()time.Time{return now}
 if err:=c.Register("a","s","fp");err!=nil{t.Fatal(err)}
 now=now.Add(61*time.Second)
 if _,_,err:=c.Identity("a");!errors.Is(err,ErrConnectionExpired){t.Fatalf("expiry: %v",err)}
 if err:=c.Register("b","s","fp");err!=nil{t.Fatal(err)}
 if err:=c.Refresh("b");err!=nil{t.Fatal(err)}
}
func TestConnectionsConcurrent(t *testing.T){
 c,_:=NewConnections(100,time.Minute)
 var wg sync.WaitGroup
 for i:=0;i<30;i++{
  wg.Add(1)
  go func(i int){
   defer wg.Done()
   id:=string(rune('A'+i))
   if err:=c.Register(id,"s","fp");err!=nil{t.Error(err);return}
   _,_,_=c.Identity(id)
   c.Close(id)
  }(i)
 }
 wg.Wait()
 if c.Count()!=0{t.Fatalf("leaked %d",c.Count())}
}
