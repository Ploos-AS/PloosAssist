//go:build linux

package screencapture

import (
 "context"
 "errors"
 "testing"
 "time"
)

type waitingPortal struct{entered chan struct{}}
func(p *waitingPortal)CreateSession(ctx context.Context)(PortalSession,error){
 close(p.entered)
 <-ctx.Done()
 return nil,ctx.Err()
}
func TestStopCancelsPendingDesktopDialog(t *testing.T){
 c:=NewPortalCapture()
 c.Approve()
 portal:=&waitingPortal{entered:make(chan struct{})}
 result:=make(chan error,1)
 go func(){result<-c.Start(context.Background(),Environment{
  SessionType:"wayland",PortalAvailable:true,PipeWireAvailable:true,
 },portal)}()
 select{case <-portal.entered:case <-time.After(time.Second):t.Fatal("portal not started")}
 done:=make(chan struct{})
 go func(){c.Stop();close(done)}()
 select{case <-done:case <-time.After(time.Second):t.Fatal("Stop blocked on portal prompt")}
 select{
 case err:=<-result:
  if !errors.Is(err,context.Canceled){t.Fatalf("expected cancellation: %v",err)}
 case <-time.After(time.Second):t.Fatal("portal request did not cancel")
 }
}
func TestDBusPortalRequiresSessionBus(t *testing.T){
 var p *DBusPortal
 _,err:=p.CreateSession(context.Background())
 if !errors.Is(err,ErrUnavailable){t.Fatalf("nil portal: %v",err)}
}
