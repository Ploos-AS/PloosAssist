package wsproto

import (
 "bytes"
 "errors"
 "testing"
)
func TestMaskedText(t *testing.T){
 payload:=[]byte("hello")
 frame:=[]byte{0x81,0x85,1,2,3,4}
 for i,b:=range payload{frame=append(frame,b^byte(i%4+1))}
 got,err:=ReadText(bytes.NewReader(frame))
 if err!=nil||string(got)!="hello"{t.Fatalf("got %q, %v",got,err)}
}
func TestUnmaskedRejected(t *testing.T){
 _,err:=ReadText(bytes.NewReader([]byte{0x81,0x01,'x'}))
 if !errors.Is(err,ErrFrame){t.Fatalf("unmasked accepted: %v",err)}
}
func TestOversizedRejected(t *testing.T){
 _,err:=ReadText(bytes.NewReader([]byte{0x81,0xfe,0x10,0x01}))
 if !errors.Is(err,ErrFrame){t.Fatalf("oversized accepted: %v",err)}
}
func TestServerTextFrame(t *testing.T){
 var buf bytes.Buffer
 if err:=WriteText(&buf,[]byte("ok"));err!=nil{t.Fatal(err)}
 if !bytes.Equal(buf.Bytes(),[]byte{0x81,2,'o','k'}){t.Fatalf("frame: %x",buf.Bytes())}
}
