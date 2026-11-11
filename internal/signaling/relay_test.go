package signaling

import (
 "errors"
 "sync"
 "testing"
)

func TestRelayConsentAndRouting(t *testing.T){
 r:=NewRelay()
 e:=Envelope{Version:1,SessionID:"s",SenderID:"host",RecipientID:"tech",Type:Offer,Payload:"sdp"}
 if err:=r.Register("s","host","tech");err!=nil{t.Fatal(err)}
 if err:=r.Authorize("host",e);!errors.Is(err,ErrUnauthorized){t.Fatal("authorized before approval")}
 if err:=r.Approve("s","attacker");!errors.Is(err,ErrUnauthorized){t.Fatal("wrong host approved")}
 if err:=r.Approve("s","host");err!=nil{t.Fatal(err)}
 if err:=r.Authorize("host",e);err!=nil{t.Fatal(err)}
 if err:=r.Authorize("attacker",e);!errors.Is(err,ErrUnauthorized){t.Fatal("sender spoof accepted")}
 e.RecipientID="other"
 if err:=r.Authorize("host",e);!errors.Is(err,ErrUnauthorized){t.Fatal("cross-peer routing accepted")}
 e.RecipientID="tech"
 if err:=r.Revoke("s","host");err!=nil{t.Fatal(err)}
 if err:=r.Authorize("host",e);!errors.Is(err,ErrUnauthorized){t.Fatal("authorized after revoke")}
}
func TestRelayConcurrentRevoke(t *testing.T){
 r:=NewRelay()
 if err:=r.Register("s","host","tech");err!=nil{t.Fatal(err)}
 if err:=r.Approve("s","host");err!=nil{t.Fatal(err)}
 e:=Envelope{Version:1,SessionID:"s",SenderID:"host",RecipientID:"tech",Type:Candidate,Payload:"ice"}
 var wg sync.WaitGroup
 for i:=0;i<20;i++{wg.Add(1);go func(){defer wg.Done();_ = r.Authorize("host",e)}()}
 if err:=r.Revoke("s","host");err!=nil{t.Fatal(err)}
 wg.Wait()
 if err:=r.Authorize("host",e);!errors.Is(err,ErrUnauthorized){t.Fatal("revoked session accepted")}
}
func TestRelayDuplicateSession(t *testing.T){
 r:=NewRelay()
 if err:=r.Register("s","h","t");err!=nil{t.Fatal(err)}
 if err:=r.Register("s","h","t");!errors.Is(err,ErrSessionExists){t.Fatal("duplicate registration accepted")}
 r.Remove("s")
 if err:=r.Register("s","h","t");err!=nil{t.Fatal(err)}
}
