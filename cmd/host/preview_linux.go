//go:build linux

package main

import (
 "bytes"
 "errors"
 "image"
 "image/png"

 "github.com/Ploos-AS/PloosAssist/internal/framebuffer"
)

// PreviewPNG is a local-only snapshot. It never transmits frames to a
// technician. A fresh copy is encoded; the native frame is never exposed.
func(a *screenShareAction)PreviewPNG()([]byte,error){
 a.mu.Lock()
 stopped:=a.stopped
 state:=a.state
 a.mu.Unlock()
 if stopped||state!="capturing"{return nil,errors.New("preview not active")}
 frame,ok:=a.local.Snapshot()
 if !ok{return nil,errors.New("no frame available")}
 png,err:=encodePreviewPNG(frame)
 clear(frame.Pixels)
 if err!=nil{return nil,err}
 a.mu.Lock()
 stopped=a.stopped || a.state!="capturing"
 a.mu.Unlock()
 if stopped{return nil,errors.New("preview stopped")}
 return png,nil
}
func encodePreviewPNG(f framebuffer.Frame)([]byte,error){
 if f.Width<=0||f.Height<=0||f.Stride<f.Width*4||
  f.Stride>framebuffer.MaxFrameBytes/f.Height||
  len(f.Pixels)!=f.Stride*f.Height{
  return nil,errors.New("invalid local frame")
 }
 img:=image.NewRGBA(image.Rect(0,0,f.Width,f.Height))
 for y:=0;y<f.Height;y++{
  row:=f.Pixels[y*f.Stride:]
  for x:=0;x<f.Width;x++{
   i:=x*4
   j:=y*img.Stride+x*4
   img.Pix[j]=row[i+2]
   img.Pix[j+1]=row[i+1]
   img.Pix[j+2]=row[i]
   img.Pix[j+3]=row[i+3]
  }
 }
 var buf bytes.Buffer
 if err:=png.Encode(&buf,img);err!=nil{return nil,err}
 return buf.Bytes(),nil
}
