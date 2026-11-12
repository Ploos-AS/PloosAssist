//go:build linux

package screencapture

import (
 "context"
 "errors"
 "sync"
 "testing"
)

type fakePortal struct{session *fakeSession;err error;creates int}
func(f *fakePortal)CreateSession(context.Context)(PortalSession,error){f.creates++;if f.err!=nil{return nil,f.err};return f.session,nil}
type fakeSession struct{mu sync.Mutex;events []string;fail string}
func(f *fakeSession)record(event string)error{
 f.mu.Lock();defer f.mu.Unlock()
 f.events=append(f.events,event)
 if event==f.fail{return ErrPortalRejected}
 return nil
}
func(f *fakeSession)SelectSources(context.Context)error{return f.record("select")}
func(f *fakeSession)Start(context.Context)error{return f.record("start")}
func(f *fakeSession)Close()error{return f.record("close")}
func(f *fakeSession)steps()[]string{f.mu.Lock();defer f.mu.Unlock();return append([]string(nil),f.events...)}

func TestPortalCaptureHandshakeAndStop(t *testing.T){
 env:=Environment{SessionType:"wayland",PortalAvailable:true,PipeWireAvailable:true}
 c:=NewPortalCapture()
 f:=&fakePortal{session:&fakeSession{}}
 if err:=c.Start(context.Background(),env,f);!errors.Is(err,ErrNotAuthorized){t.Fatalf("without consent: %v",err)}
 if f.creates!=0{t.Fatal("portal contacted before consent")}
 c.Approve()
 if err:=c.Start(context.Background(),env,f);err!=nil{t.Fatal(err)}
 c.Stop()
 got:=f.session.steps()
 if len(got)!=3||got[0]!="select"||got[1]!="start"||got[2]!="close"{t.Fatalf("portal order: %v",got)}
 if err:=c.Start(context.Background(),env,f);!errors.Is(err,ErrStopped){t.Fatalf("resumed after Stop: %v",err)}
}

func TestPortalCaptureFailureClosesSession(t *testing.T){
 env:=Environment{SessionType:"wayland",PortalAvailable:true,PipeWireAvailable:true}
 for _,step:=range []string{"select","start"}{
  t.Run(step,func(t *testing.T){
   c:=NewPortalCapture()
   c.Approve()
   f:=&fakePortal{session:&fakeSession{fail:step}}
   if err:=c.Start(context.Background(),env,f);!errors.Is(err,ErrPortalRejected){t.Fatalf("portal error: %v",err)}
   got:=f.session.steps()
   if got[len(got)-1]!="close"{t.Fatalf("session not closed: %v",got)}
  })
 }
}

func TestPortalCaptureUnavailableFailsBeforePortal(t *testing.T){
 c:=NewPortalCapture()
 c.Approve()
 f:=&fakePortal{session:&fakeSession{}}
 if err:=c.Start(context.Background(),Environment{SessionType:"x11"},f);!errors.Is(err,ErrUnavailable){t.Fatal(err)}
 if f.creates!=0{t.Fatal("unexpected portal request")}
 if err:=c.Start(context.Background(),Environment{SessionType:"wayland",PortalAvailable:true,PipeWireAvailable:true},nil);!errors.Is(err,ErrNotImplemented){t.Fatal(err)}
}
