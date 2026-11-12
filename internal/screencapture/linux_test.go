//go:build linux

package screencapture

import (
 "context"
 "errors"
 "testing"
)

func TestCaptureRequiresExplicitConsent(t *testing.T){
 c:=NewController()
 ready:=Environment{SessionType:"wayland",PortalAvailable:true,PipeWireAvailable:true}
 if err:=c.Start(context.Background(),ready);!errors.Is(err,ErrNotAuthorized){t.Fatalf("unapproved capture: %v",err)}
 c.Approve()
 if err:=c.Start(context.Background(),Environment{SessionType:"x11",PortalAvailable:true,PipeWireAvailable:true});!errors.Is(err,ErrUnavailable){t.Fatalf("unsupported backend: %v",err)}
 if err:=c.Start(context.Background(),ready);!errors.Is(err,ErrNotImplemented){t.Fatalf("unexpected capture implementation: %v",err)}
 c.Stop()
 c.Approve()
 if err:=c.Start(context.Background(),ready);!errors.Is(err,ErrStopped){t.Fatalf("capture resumed after stop: %v",err)}
}

func TestCaptureRespectsCancelledContext(t *testing.T){
 c:=NewController()
 c.Approve()
 ctx,cancel:=context.WithCancel(context.Background())
 cancel()
 err:=c.Start(ctx,Environment{SessionType:"wayland",PortalAvailable:true,PipeWireAvailable:true})
 if !errors.Is(err,context.Canceled){t.Fatalf("cancelled context: %v",err)}
}
