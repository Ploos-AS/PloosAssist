//go:build linux

package screencapture

import (
 "context"
 "errors"
 "os"
 "sync"

 "github.com/Ploos-AS/PloosAssist/internal/framebuffer"
)

// ErrConsumerUnavailable indicates no native PipeWire video backend is
// installed. A raw PipeWire FD is NOT a raw video byte stream.
var ErrConsumerUnavailable=errors.New("native PipeWire video consumer not installed")

// VideoConsumer represents a future native PipeWire video decoder. It must
// connect through the portal-provided FD and select ONLY an approved node.
type VideoConsumer interface{
 Run(context.Context,*os.File,uint32,*framebuffer.Latest)error
}

// LocalCapture binds an approved portal session, its remote FD and a
// video consumer to a local-only bounded frame buffer.
type LocalCapture struct{
 mu sync.Mutex
 stopped bool
 active bool
 buffer *framebuffer.Latest
 cancel context.CancelFunc
}
func NewLocalCapture()*LocalCapture{
 return &LocalCapture{buffer:framebuffer.NewLatest()}
}
func(c *LocalCapture)Snapshot()(framebuffer.Frame,bool){return c.buffer.Snapshot()}
func(c *LocalCapture)Stop(){
 c.mu.Lock()
 c.stopped=true
 cancel:=c.cancel
 c.mu.Unlock()
 if cancel!=nil{cancel()}
 c.buffer.Close()
}
// Run consumes only the first explicitly approved stream. The FD is always
// closed, including when validation or consumer startup fails.
func(c *LocalCapture)Run(ctx context.Context,session *dbusPortalSession,consumer VideoConsumer)error{
 if session==nil{return ErrNotAuthorized}
 streams,err:=session.Streams()
 if err!=nil{return err}
 if len(streams)==0{return ErrNotAuthorized}
 c.mu.Lock()
 if c.stopped{c.mu.Unlock();return ErrStopped}
 if c.active{c.mu.Unlock();return ErrPortalRejected}
 if consumer==nil{c.mu.Unlock();return ErrConsumerUnavailable}
 runCtx,cancel:=context.WithCancel(ctx)
 c.cancel=cancel
 c.active=true
 c.mu.Unlock()
 defer func(){
  cancel()
  c.mu.Lock()
  c.cancel=nil
  c.active=false
  c.mu.Unlock()
  c.buffer.Close()
 }()
 fd,err:=session.OpenPipeWireRemote(runCtx)
 if err!=nil{return err}
 defer fd.Close()
 return consumer.Run(runCtx,fd,streams[0].NodeID,c.buffer)
}
