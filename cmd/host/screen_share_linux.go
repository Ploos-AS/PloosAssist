//go:build linux

package main

import (
 "context"
 "errors"
 "sync"

 "github.com/Ploos-AS/PloosAssist/internal/screencapture"
)

type screenShareAction struct{
 mu sync.Mutex
 capture *screencapture.PortalCapture
 requested bool
 stopped bool
 state string
}
func newScreenShareAction()*screenShareAction{
 return &screenShareAction{capture:screencapture.NewPortalCapture(),state:"idle"}
}
func(a *screenShareAction)Status()string{
 a.mu.Lock();defer a.mu.Unlock()
 return a.state
}
func(a *screenShareAction)Start(ctx context.Context)error{
 a.mu.Lock()
 if a.stopped{a.mu.Unlock();return errors.New("session stopped")}
 if a.requested{a.mu.Unlock();return errors.New("screen share already requested")}
 if err:=ctx.Err();err!=nil{a.mu.Unlock();return err}
 a.requested=true
 a.state="waiting"
 a.capture.Approve()
 a.mu.Unlock()
 portal,err:=screencapture.NewDBusPortal()
 if err!=nil{
  a.mu.Lock();if !a.stopped{a.state="error"};a.mu.Unlock()
  return err
 }
 go func(){
  defer portal.Close()
  err:=a.capture.Start(ctx,screencapture.Probe(),portal)
  a.mu.Lock()
  defer a.mu.Unlock()
  if a.stopped{return}
  switch{
  case err==nil:a.state="approved"
  case errors.Is(err,screencapture.ErrPortalResponse),errors.Is(err,screencapture.ErrPortalRejected):a.state="denied"
  default:a.state="error"
  }
 }()
 return nil
}
func(a *screenShareAction)Stop(){
 a.mu.Lock()
 a.stopped=true
 a.state="stopped"
 a.mu.Unlock()
 a.capture.Stop()
}
