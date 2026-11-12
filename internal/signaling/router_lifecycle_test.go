package signaling

import (
 "testing"
 "time"
)
func TestRouterDetachAfterConnectionIdentityClosed(t *testing.T){
 connections,_:=NewConnections(4,time.Minute)
 consent,_:=NewConsentPolicy(4)
 router,_:=NewRouter(connections,consent)
 _=connections.Register("old","session","same-peer")
 if _,err:=router.Attach("old");err!=nil{t.Fatal(err)}
 connections.Close("old")
 router.Detach("old")
 _=connections.Register("new","session","same-peer")
 if _,err:=router.Attach("new");err!=nil{t.Fatalf("stale destination blocked replacement: %v",err)}
 router.Detach("old")
 if len(router.destinations["session"])!=1{t.Fatal("stale detach removed replacement")}
 router.Detach("new")
 if len(router.destinations)!=0||len(router.owners)!=0{t.Fatal("router retained disconnected destinations")}
}
func TestRouterCloseSessionClearsOwners(t *testing.T){
 connections,_:=NewConnections(4,time.Minute)
 consent,_:=NewConsentPolicy(4)
 router,_:=NewRouter(connections,consent)
 _=connections.Register("one","session","one-fp")
 if _,err:=router.Attach("one");err!=nil{t.Fatal(err)}
 router.CloseSession("session")
 if len(router.owners)!=0||len(router.destinations)!=0{t.Fatal("session close retained destinations")}
}
