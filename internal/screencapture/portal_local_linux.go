//go:build linux

package screencapture

import (
 "context"
 "errors"
)

// CaptureLocal first obtains an explicit desktop portal grant, then starts
// the native local-only consumer on the very same authorized portal session.
// Stop must be called on BOTH controllers by their host owner.
func(c *PortalCapture)CaptureLocal(ctx context.Context,env Environment,portal Portal,consumer VideoConsumer,local *LocalCapture)error{
 if local==nil||consumer==nil{return ErrConsumerUnavailable}
 if err:=c.Start(ctx,env,portal);err!=nil{return err}
 c.mu.Lock()
 session:=c.session
 stopped:=c.stopped
 c.mu.Unlock()
 if stopped{return ErrStopped}
 native,ok:=session.(*dbusPortalSession)
 if !ok{return errors.New("local capture requires native desktop portal session")}
 return local.Run(ctx,native,consumer)
}
