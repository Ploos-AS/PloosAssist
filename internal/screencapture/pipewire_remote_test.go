//go:build linux

package screencapture

import (
 "context"
 "errors"
 "testing"
)

func TestPipeWireRemoteRequiresPortalGrant(t *testing.T){
 s:=&dbusPortalSession{}
 if _,err:=s.OpenPipeWireRemote(context.Background());!errors.Is(err,ErrNotAuthorized){
  t.Fatalf("remote before Start grant: %v",err)
 }
 s.closed=true
 if _,err:=s.OpenPipeWireRemote(context.Background());!errors.Is(err,ErrStopped){
  t.Fatalf("remote after Close: %v",err)
 }
}
func TestPipeWireRemoteCancelledBeforeDBus(t *testing.T){
 s:=&dbusPortalSession{started:true}
 ctx,cancel:=context.WithCancel(context.Background())
 cancel()
 if _,err:=s.OpenPipeWireRemote(ctx);!errors.Is(err,context.Canceled){
  t.Fatalf("cancelled request: %v",err)
 }
}
