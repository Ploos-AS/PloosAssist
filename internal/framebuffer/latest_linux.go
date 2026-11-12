//go:build linux

// Package framebuffer provides a bounded, local-only staging area for
// future portal-authorized PipeWire video frames. It does not capture them.
package framebuffer

import (
 "errors"
 "sync"
)

var (
 ErrClosed=errors.New("frame buffer closed")
 ErrInvalidFrame=errors.New("invalid frame")
 ErrTooLarge=errors.New("frame exceeds size limit")
)
const MaxFrameBytes=16<<20 // 16 MiB, hard cap

// Frame is a copy of one uncompressed BGRA frame in local process memory.
// A real PipeWire adapter must validate buffer format and stride first.
type Frame struct {
 Width int
 Height int
 Stride int
 Pixels []byte
}
func validate(f Frame)error{
 if f.Width<=0||f.Height<=0||f.Stride<=0{return ErrInvalidFrame}
 if f.Width>8192||f.Height>8192{return ErrInvalidFrame}
 if f.Width>f.Stride/4{return ErrInvalidFrame}
 if f.Height>MaxFrameBytes/f.Stride{return ErrTooLarge}
 expected:=f.Stride*f.Height
 if expected!=len(f.Pixels){return ErrInvalidFrame}
 return nil
}
// Latest holds at most one frame and never exposes its backing memory.
type Latest struct{
 mu sync.Mutex
 closed bool
 frame Frame
 present bool
}
func NewLatest()*Latest{return &Latest{}}
func(s *Latest)Put(f Frame)error{
 if err:=validate(f);err!=nil{return err}
 copyPixels:=make([]byte,len(f.Pixels))
 copy(copyPixels,f.Pixels)
 s.mu.Lock()
 defer s.mu.Unlock()
 if s.closed{clear(copyPixels);return ErrClosed}
 if s.present{clear(s.frame.Pixels)}
 s.frame=Frame{Width:f.Width,Height:f.Height,Stride:f.Stride,Pixels:copyPixels}
 s.present=true
 return nil
}
func(s *Latest)Snapshot()(Frame,bool){
 s.mu.Lock()
 defer s.mu.Unlock()
 if s.closed||!s.present{return Frame{},false}
 pixels:=append([]byte(nil),s.frame.Pixels...)
 return Frame{Width:s.frame.Width,Height:s.frame.Height,Stride:s.frame.Stride,Pixels:pixels},true
}
func(s *Latest)Close(){
 s.mu.Lock()
 defer s.mu.Unlock()
 if s.present{clear(s.frame.Pixels)}
 s.frame=Frame{}
 s.present=false
 s.closed=true
}
