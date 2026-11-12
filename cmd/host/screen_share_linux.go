//go:build linux

package main

import (
 "context"
 "errors"
 "sync"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/screencapture"
)

type screenShareAction struct{
 mu sync.Mutex
 capture *screencapture.PortalCapture
 local *screencapture.LocalCapture
 requested bool
 stopped bool
 state string
 frames uint64
}
func newScreenShareAction()*screenShareAction{
 return &screenShareAction{
  capture:screencapture.NewPortalCapture(),
  local:screencapture.NewLocalCapture(),
  state:"idle",
 }
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
  consumer:=localVideoConsumer()
  if consumer==nil{
   // Default builds retain the existing permission-only prototype.
   err:=a.capture.Start(ctx,screencapture.Probe(),portal)
   a.setOutcome(err,false)
   return
  }
  finished:=make(chan struct{})
  go func(){
   ticker:=time.NewTicker(100*time.Millisecond)
   defer ticker.Stop()
   for{
    select{
    case <-finished:return
    case <-ctx.Done():return
    case <-ticker.C:
     if _,ok:=a.local.Snapshot();ok{
      a.mu.Lock()
      if !a.stopped{a.state="capturing"}
      a.mu.Unlock()
     }
    }
   }
  }()
  err:=a.capture.CaptureLocal(ctx,screencapture.Probe(),portal,consumer,a.local)
  close(finished)
  a.setOutcome(err,true)
 }()
 return nil
}
func(a *screenShareAction)setOutcome(err error,native bool){
 a.mu.Lock()
 defer a.mu.Unlock()
 if a.stopped{return}
 switch{
 case err==nil && !native:a.state="approved"
 case err==nil && native:a.state="finished"
 case errors.Is(err,screencapture.ErrPortalResponse),errors.Is(err,screencapture.ErrPortalRejected):a.state="denied"
 default:a.state="error"
 }
}
func(a *screenShareAction)Stop(){
 a.mu.Lock()
 a.stopped=true
 a.state="stopped"
 a.mu.Unlock()
 a.local.Stop()
 a.capture.Stop()
}
