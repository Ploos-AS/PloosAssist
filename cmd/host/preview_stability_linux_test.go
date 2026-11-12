//go:build linux

package main

import (
 "bytes"
 "image/png"
 "sync"
 "testing"

 "github.com/Ploos-AS/PloosAssist/internal/framebuffer"
)

func TestPreviewCapturingOnlyAndStop(t *testing.T){
 a:=newScreenShareAction()
 frame:=framebuffer.Frame{Width:2,Height:1,Stride:8,Pixels:[]byte{1,2,3,255,4,5,6,255}}
 if err:=a.local.Put(frame);err!=nil{t.Fatal(err)}
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview before capture state")}
 a.mu.Lock();a.state="capturing";a.mu.Unlock()
 data,err:=a.PreviewPNG()
 if err!=nil{t.Fatal(err)}
 decoded,err:=png.Decode(bytes.NewReader(data))
 if err!=nil{t.Fatal(err)}
 if decoded.Bounds().Dx()!=2{t.Fatal("unexpected preview width")}
 a.Stop()
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview after Stop")}
}
func TestConcurrentPreviewStop(t *testing.T){
 a:=newScreenShareAction()
 if err:=a.local.Put(framebuffer.Frame{Width:1,Height:1,Stride:4,Pixels:[]byte{1,2,3,255}});err!=nil{t.Fatal(err)}
 a.mu.Lock();a.state="capturing";a.mu.Unlock()
 var wg sync.WaitGroup
 for i:=0;i<8;i++{
  wg.Add(1)
  go func(){
   defer wg.Done()
   for j:=0;j<30;j++{_,_=a.PreviewPNG()}
  }()
 }
 a.Stop()
 wg.Wait()
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview available after Stop")}
}
