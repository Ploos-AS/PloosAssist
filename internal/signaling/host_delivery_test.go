package signaling

import (
 "errors"
 "testing"
 "time"
)
func setupHostDelivery(t *testing.T)(*HostDelivery,*ConsentPolicy,*Connections,ApprovalTicket){
 t.Helper()
 connections,_:=NewConnections(4,time.Minute)
 consent,_:=NewConsentPolicy(3)
 coordinator,_:=NewCoordinator(connections,consent)
 approvals,_:=NewApprovalRequests(3,connections,coordinator)
 events,_:=NewApprovalEvents(3,approvals)
 delivery,_:=NewHostDelivery(3,connections,approvals,events)
 _=connections.Register("host","s","host-fp")
 _=connections.Register("tech","s","tech-fp")
 _=connections.Register("other","s","other-fp")
 request,err:=approvals.Request("s","host","tech",time.Minute)
 if err!=nil{t.Fatal(err)}
 ticket,err:=delivery.Queue(request,"host")
 if err!=nil{t.Fatal(err)}
 return delivery,consent,connections,ticket
}
func TestHostDeliveryRequiresOriginalConnection(t *testing.T){
 delivery,consent,_,ticket:=setupHostDelivery(t)
 if _,err:=delivery.Take("other");!errors.Is(err,ErrHostDelivery){t.Fatalf("other take: %v",err)}
 if err:=delivery.Decide("other",ticket.ID,true);!errors.Is(err,ErrHostDelivery){t.Fatalf("other approve: %v",err)}
 got,err:=delivery.Take("host")
 if err!=nil||got.ID!=ticket.ID{t.Fatalf("host take: %+v %v",got,err)}
 if err:=delivery.Decide("host",ticket.ID,true);err!=nil{t.Fatal(err)}
 if !consent.Allowed("s","host-fp","tech-fp"){t.Fatal("consent missing")}
 if err:=delivery.Decide("host",ticket.ID,true);!errors.Is(err,ErrHostDelivery){t.Fatalf("replay: %v",err)}
}
func TestHostDeliveryDisconnectCancelsRequest(t *testing.T){
 delivery,consent,connections,ticket:=setupHostDelivery(t)
 delivery.Disconnect("host")
 connections.Close("host")
 if err:=delivery.Decide("host",ticket.ID,true);!errors.Is(err,ErrHostDelivery){t.Fatalf("after disconnect: %v",err)}
 if consent.Allowed("s","host-fp","tech-fp"){t.Fatal("disconnect approved request")}
}
func TestHostDeliveryDenial(t *testing.T){
 delivery,consent,_,ticket:=setupHostDelivery(t)
 if err:=delivery.Decide("host",ticket.ID,false);err!=nil{t.Fatal(err)}
 if consent.Allowed("s","host-fp","tech-fp"){t.Fatal("denial granted consent")}
}
