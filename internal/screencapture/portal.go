//go:build linux

package screencapture

import (
 "context"
 "errors"
 "sync"
)

var ErrPortalRejected=errors.New("desktop portal request rejected")

// PortalSession is an opaque desktop-portal session handle. It must never
// be derived from a caller-provided path or used to bypass desktop prompts.
type PortalSession interface {
 SelectSources(context.Context)error
 Start(context.Context)error
 Close()error
}

// Portal is an abstraction over the desktop's ScreenCast portal. A future
// D-Bus implementation must validate response handles and await the
// asynchronous Response signal at each step.
type Portal interface {
 CreateSession(context.Context)(PortalSession,error)
}

// PortalCapture is a one-shot, explicitly authorized portal lifecycle.
// The controller never returns PipeWire frames or remote control access.
type PortalCapture struct {
 mu sync.Mutex
 approved bool
 stopped bool
 started bool
 session PortalSession
 cancel context.CancelFunc
}
func NewPortalCapture()*PortalCapture{return &PortalCapture{}}
func(c *PortalCapture)Approve(){
 c.mu.Lock();defer c.mu.Unlock()
 if !c.stopped{c.approved=true}
}
func(c *PortalCapture)Stop(){
 c.mu.Lock()
 c.stopped=true
 c.approved=false
 session:=c.session
 cancel:=c.cancel
 c.session=nil
 c.mu.Unlock()
 if cancel!=nil{cancel()}
 if session!=nil{_ =session.Close()}
}
func(c *PortalCapture)Start(ctx context.Context,env Environment,portal Portal)error{
 c.mu.Lock()
 if c.stopped{c.mu.Unlock();return ErrStopped}
 if !c.approved{c.mu.Unlock();return ErrNotAuthorized}
 if c.started{c.mu.Unlock();return ErrPortalRejected}
 if err:=ctx.Err();err!=nil{c.mu.Unlock();return err}
 if env.SessionType!="wayland"||!env.PortalAvailable||!env.PipeWireAvailable{c.mu.Unlock();return ErrUnavailable}
 if portal==nil{c.mu.Unlock();return ErrNotImplemented}
 // Never hold the mutex while waiting for a user's desktop portal dialog.
 // Stop must be able to cancel an in-flight request immediately.
 c.started=true
 runCtx,cancel:=context.WithCancel(ctx)
 c.cancel=cancel
 c.mu.Unlock()
 defer func(){
  cancel()
  c.mu.Lock()
  c.cancel=nil
  c.mu.Unlock()
 }()
 session,err:=portal.CreateSession(runCtx)
 if err!=nil{return err}
 if session==nil{return ErrPortalRejected}
 c.mu.Lock()
 if c.stopped{c.mu.Unlock();_ =session.Close();return ErrStopped}
 c.session=session
 c.mu.Unlock()
 if err:=session.SelectSources(runCtx);err!=nil{_ =session.Close();return err}
 if err:=session.Start(runCtx);err!=nil{_ =session.Close();return err}
 if err:=runCtx.Err();err!=nil{_ =session.Close();return err}
 c.mu.Lock()
 stopped:=c.stopped
 c.mu.Unlock()
 if stopped{_ =session.Close();return ErrStopped}
 return nil
}
