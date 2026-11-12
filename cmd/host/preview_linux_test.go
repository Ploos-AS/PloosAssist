//go:build linux

package main

import (
 "bytes"
 "image/png"
 "testing"

 "github.com/Ploos-AS/PloosAssist/internal/framebuffer"
)

func TestEncodePreviewPNGSwapsBGRA(t *testing.T){
 data,err:=encodePreviewPNG(framebuffer.Frame{
  Width:1,Height:1,Stride:4,Pixels:[]byte{10,20,30,255},
 })
 if err!=nil{t.Fatal(err)}
 img,err:=png.Decode(bytes.NewReader(data))
 if err!=nil{t.Fatal(err)}
 r,g,b,a:=img.At(0,0).RGBA()
 if r>>8!=30||g>>8!=20||b>>8!=10||a>>8!=255{
  t.Fatalf("unexpected color: %d %d %d %d",r>>8,g>>8,b>>8,a>>8)
 }
}
func TestPreviewUnavailableBeforeCaptureAndAfterStop(t *testing.T){
 a:=newScreenShareAction()
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview before consent")}
 a.Stop()
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview after stop")}
}
func TestEncodePreviewRejectsMalformedFrame(t *testing.T){
 if _,err:=encodePreviewPNG(framebuffer.Frame{Width:2,Height:1,Stride:4,Pixels:make([]byte,4)});err==nil{
  t.Fatal("accepted short stride")
 }
}
