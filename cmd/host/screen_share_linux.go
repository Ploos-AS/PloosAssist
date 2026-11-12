//go:build linux

package main

import (
 "context"
 "errors"
 "sync"

 "github.com/Ploos-AS/PloosAssist/internal/screencapture"
)

// screenShareAction owns a single local desktop portal request. It never
// captures pixels, opens PipeWire or enables remote input.
type screenShareAction struct{
 mu sync.Mutex
 capture *screencapture.PortalCapture
 requested bool
}
func newScreenShareAction()*screenShareAction{
 return &screenShareAction{capture:screencapture.NewPortalCapture()}
}
func(a *screenShareAction)Start(ctx context.Context)error{
 a.mu.Lock()
 if a.requested{a.mu.Unlock();return errors.New("screen share already requested")}
 if err:=ctx.Err();err!=nil{a.mu.Unlock();return err}
 a.requested=true
 a.capture.Approve()
 a.mu.Unlock()
 portal,err:=screencapture.NewDBusPortal()
 if err!=nil{return err}
 go func(){
  defer portal.Close()
  // Start waits for the desktop portal's interactive user selection.
  // No media stream is opened after the permission grant.
  _=a.capture.Start(ctx,screencapture.Probe(),portal)
 }()
 return nil
}
func(a *screenShareAction)Stop(){a.capture.Stop()}
