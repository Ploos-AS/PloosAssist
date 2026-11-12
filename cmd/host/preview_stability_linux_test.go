//go:build linux

package main

import (
 "sync"
 "testing"
)

func TestPreviewCapturingStateWithoutFrameFailsClosed(t *testing.T){
 a:=newScreenShareAction()
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview before capture state")}
 a.mu.Lock();a.state="capturing";a.mu.Unlock()
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview with no captured frame")}
 a.Stop()
 if _,err:=a.PreviewPNG();err==nil{t.Fatal("preview after Stop")}
}
func TestConcurrentPreviewStop(t *testing.T){
 a:=newScreenShareAction()
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
