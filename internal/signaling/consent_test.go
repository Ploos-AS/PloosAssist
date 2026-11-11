package signaling

import (
 "errors"
 "sync"
 "testing"
 "time"
)
func TestConsentSpecificTechnician(t *testing.T){
 p,_:=NewConsentPolicy(2)
 if err:=p.Create("s","host","tech",time.Now().Add(time.Minute));err!=nil{t.Fatal(err)}
 if p.Allowed("s","host","tech"){t.Fatal("allowed without consent")}
 if err:=p.Approve("s","host","other");!errors.Is(err,ErrConsentDenied){t.Fatal("wrong technician approved")}
 if err:=p.Approve("s","attacker","tech");!errors.Is(err,ErrConsentDenied){t.Fatal("wrong host approved")}
 if err:=p.Approve("s","host","tech");err!=nil{t.Fatal(err)}
 if !p.Allowed("s","host","tech")||!p.Allowed("s","tech","host"){t.Fatal("approved peers blocked")}
 if p.Allowed("s","tech","other")||p.Allowed("s","other","host"){t.Fatal("third-party peer allowed")}
 if err:=p.Revoke("s","host");err!=nil{t.Fatal(err)}
 if p.Allowed("s","host","tech"){t.Fatal("allowed after revoke")}
}
func TestConsentExpiryAndCapacity(t *testing.T){
 p,_:=NewConsentPolicy(1)
 now:=time.Now()
 p.now=func()time.Time{return now}
 if err:=p.Create("s","host","tech",now.Add(time.Second));err!=nil{t.Fatal(err)}
 if err:=p.Create("s2","h2","t2",now.Add(time.Minute));!errors.Is(err,ErrConnectionLimit){t.Fatal("capacity not enforced")}
 if err:=p.Approve("s","host","tech");err!=nil{t.Fatal(err)}
 now=now.Add(2*time.Second)
 if p.Allowed("s","host","tech"){t.Fatal("expired consent allowed")}
 if err:=p.Create("s2","h2","t2",now.Add(time.Minute));err!=nil{t.Fatal(err)}
}
func TestConsentConcurrentChecksAndRevocation(t *testing.T){
 p,_:=NewConsentPolicy(1)
 if err:=p.Create("s","host","tech",time.Now().Add(time.Minute));err!=nil{t.Fatal(err)}
 if err:=p.Approve("s","host","tech");err!=nil{t.Fatal(err)}
 var wg sync.WaitGroup
 for i:=0;i<30;i++{wg.Add(1);go func(){defer wg.Done();_=p.Allowed("s","tech","host")}()}
 if err:=p.Revoke("s","host");err!=nil{t.Fatal(err)}
 wg.Wait()
 if p.Allowed("s","tech","host"){t.Fatal("revoked session allowed")}
}
