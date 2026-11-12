package signaling

import (
 "errors"
 "testing"
 "time"
)

func TestRouterRequiresVerifiedIdentityAndLiveConsent(t *testing.T){
 connections,_:=NewConnections(4,time.Minute)
 consent,_:=NewConsentPolicy(2)
 router,_:=NewRouter(connections,consent)
 if err:=connections.Register("host-conn","session","host-fp");err!=nil{t.Fatal(err)}
 if err:=connections.Register("tech-conn","session","tech-fp");err!=nil{t.Fatal(err)}
 if _,err:=router.Attach("host-conn");err!=nil{t.Fatal(err)}
 inbox,err:=router.Attach("tech-conn");if err!=nil{t.Fatal(err)}
 message:=Envelope{Version:1,SessionID:"session",SenderID:"host-fp",RecipientID:"tech-fp",Type:Offer,Payload:"offer"}
 if err:=router.Route("host-conn",message);!errors.Is(err,ErrConsentDenied){t.Fatalf("missing consent: %v",err)}
 if err:=consent.Create("session","host-fp","tech-fp",time.Now().Add(time.Minute));err!=nil{t.Fatal(err)}
 if err:=consent.Approve("session","host-fp","tech-fp");err!=nil{t.Fatal(err)}
 if err:=router.Route("host-conn",message);err!=nil{t.Fatal(err)}
 select{case got:=<-inbox:if got.Payload!="offer"{t.Fatalf("payload: %s",got.Payload)};default:t.Fatal("missing delivery")}
 message.SenderID="tech-fp"
 if err:=router.Route("host-conn",message);!errors.Is(err,ErrUnauthorized){t.Fatalf("spoofing: %v",err)}
 message.SenderID="host-fp"
 if err:=consent.Revoke("session","host-fp");err!=nil{t.Fatal(err)}
 if err:=router.Route("host-conn",message);!errors.Is(err,ErrConsentDenied){t.Fatalf("revocation: %v",err)}
 router.Detach("tech-conn")
 connections.Close("tech-conn")
 if err:=consent.Approve("session","host-fp","tech-fp");err!=nil{t.Fatal(err)}
 if err:=router.Route("host-conn",message);!errors.Is(err,ErrPeerOffline){t.Fatalf("offline peer: %v",err)}
}
func TestRouterBoundedQueue(t *testing.T){
 connections,_:=NewConnections(2,time.Minute)
 consent,_:=NewConsentPolicy(1)
 router,_:=NewRouter(connections,consent)
 _=connections.Register("host","s","h")
 _=connections.Register("tech","s","t")
 _,_=router.Attach("tech")
 _=consent.Create("s","h","t",time.Now().Add(time.Minute))
 _=consent.Approve("s","h","t")
 e:=Envelope{Version:1,SessionID:"s",SenderID:"h",RecipientID:"t",Type:Candidate,Payload:"candidate"}
 for i:=0;i<16;i++{if err:=router.Route("host",e);err!=nil{t.Fatal(err)}}
 if err:=router.Route("host",e);!errors.Is(err,ErrConnectionLimit){t.Fatalf("backpressure: %v",err)}
}
