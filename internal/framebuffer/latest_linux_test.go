//go:build linux

package framebuffer

import (
 "bytes"
 "errors"
 "sync"
 "testing"
)

func TestFrameCopyIsolationAndClose(t *testing.T){
 b:=NewLatest()
 source:=Frame{Width:2,Height:1,Stride:8,Pixels:[]byte{1,2,3,4,5,6,7,8}}
 if err:=b.Put(source);err!=nil{t.Fatal(err)}
 source.Pixels[0]=99
 snapshot,ok:=b.Snapshot()
 if !ok||snapshot.Pixels[0]!=1{t.Fatalf("source mutated buffer: %v",snapshot.Pixels)}
 snapshot.Pixels[0]=88
 second,ok:=b.Snapshot()
 if !ok||second.Pixels[0]!=1{t.Fatal("snapshot exposed backing memory")}
 b.Close()
 if _,ok:=b.Snapshot();ok{t.Fatal("frame visible after Close")}
 if err:=b.Put(source);!errors.Is(err,ErrClosed){t.Fatalf("accepted after Close: %v",err)}
}
func TestInvalidAndOversizedFrames(t *testing.T){
 b:=NewLatest()
 cases:=[]Frame{
  {Width:1,Height:1,Stride:4,Pixels:[]byte{1}},
  {Width:8193,Height:1,Stride:32772,Pixels:make([]byte,32772)},
  {Width:1,Height:1,Stride:0},
  {Width:2,Height:1,Stride:4,Pixels:make([]byte,4)},
  {Width:1,Height:2,Stride:MaxFrameBytes,Pixels:nil},
 }
 for i,f:=range cases{
  if err:=b.Put(f);err==nil{t.Fatalf("accepted invalid frame %d",i)}
 }
}
func TestLatestReplacesFrame(t *testing.T){
 b:=NewLatest()
 for _,v:=range []byte{1,2,3}{
  if err:=b.Put(Frame{Width:1,Height:1,Stride:4,Pixels:bytes.Repeat([]byte{v},4)});err!=nil{t.Fatal(err)}
 }
 f,ok:=b.Snapshot()
 if !ok||f.Pixels[0]!=3{t.Fatal("did not keep latest frame")}
}
func TestConcurrentPutSnapshotClose(t *testing.T){
 b:=NewLatest()
 var wg sync.WaitGroup
 for i:=0;i<4;i++{
  wg.Add(1)
  go func(){
   defer wg.Done()
   for j:=0;j<100;j++{
    _=b.Put(Frame{Width:1,Height:1,Stride:4,Pixels:[]byte{1,2,3,4}})
    _,_=b.Snapshot()
   }
  }()
 }
 b.Close()
 wg.Wait()
}
