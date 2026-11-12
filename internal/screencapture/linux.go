//go:build linux

// Package screencapture defines a fail-closed Linux screen-share lifecycle.
// No frames or remote-input events are produced by this foundation.
package screencapture

import (
 "context"
 "errors"
 "os"
 "os/exec"
 "strings"
 "sync"
)

var (
 ErrUnavailable=errors.New("desktop portal screencast prerequisites unavailable")
 ErrNotAuthorized=errors.New("screen sharing requires explicit session consent")
 ErrNotImplemented=errors.New("portal screencast negotiation is not implemented")
 ErrStopped=errors.New("screen sharing session stopped")
)

// Environment describes local prerequisites; it is not authorization.
type Environment struct {
 SessionType string
 PortalAvailable bool
 PipeWireAvailable bool
}
func Probe() Environment {
 env:=Environment{SessionType:strings.ToLower(os.Getenv("XDG_SESSION_TYPE"))}
 _,portal:=exec.LookPath("xdg-desktop-portal")
 _,pipewire:=exec.LookPath("pipewire")
 env.PortalAvailable=portal==nil
 env.PipeWireAvailable=pipewire==nil
 return env
}

// Controller holds an explicit consent flag for a single, locally owned
// capture attempt. It never grants server-side or remote-control permissions.
type Controller struct {
 mu sync.Mutex
 approved bool
 stopped bool
}
func NewController()*Controller{return &Controller{}}
func(c *Controller)Approve(){
 c.mu.Lock();defer c.mu.Unlock()
 if !c.stopped{c.approved=true}
}
func(c *Controller)Stop(){
 c.mu.Lock();defer c.mu.Unlock()
 c.approved=false
 c.stopped=true
}
func(c *Controller)Start(ctx context.Context,env Environment)error{
 c.mu.Lock()
 defer c.mu.Unlock()
 if c.stopped{return ErrStopped}
 if !c.approved{return ErrNotAuthorized}
 if err:=ctx.Err();err!=nil{return err}
 if env.SessionType!="wayland"||!env.PortalAvailable||!env.PipeWireAvailable{return ErrUnavailable}
 // A real implementation must use org.freedesktop.portal.ScreenCast
 // CreateSession -> SelectSources -> Start and wait for user selection,
 // then consume PipeWire frames. Never silently capture the desktop.
 return ErrNotImplemented
}
