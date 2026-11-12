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
 session PortalSession
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
 c.session=nil
 c.mu.Unlock()
 if session!=nil{_ =session.Close()}
}
func(c *PortalCapture)Start(ctx context.Context,env Environment,portal Portal)error{
 c.mu.Lock()
 defer c.mu.Unlock()
 if c.stopped{return ErrStopped}
 if !c.approved{return ErrNotAuthorized}
 if err:=ctx.Err();err!=nil{return err}
 if env.SessionType!="wayland"||!env.PortalAvailable||!env.PipeWireAvailable{return ErrUnavailable}
 if portal==nil{return ErrNotImplemented}
 session,err:=portal.CreateSession(ctx)
 if err!=nil{return err}
 if session==nil{return ErrPortalRejected}
 // Retain the lock during the entire consent handshake so Stop cannot
 // race with the transition into an authorized capture session.
 c.session=session
 if err:=session.SelectSources(ctx);err!=nil{_ =session.Close();c.session=nil;return err}
 if err:=session.Start(ctx);err!=nil{_ =session.Close();c.session=nil;return err}
 if err:=ctx.Err();err!=nil{_ =session.Close();c.session=nil;return err}
 // No PipeWire remote is opened and no frames leave the machine.
 return nil
}
