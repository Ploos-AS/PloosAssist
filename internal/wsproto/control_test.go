package wsproto

import (
 "bytes"
 "errors"
 "testing"
)

func masked(op byte,final bool,payload []byte)[]byte{
 first:=op
 if final{first|=0x80}
 frame:=[]byte{first,0x80|byte(len(payload)),1,2,3,4}
 for i,b:=range payload{frame=append(frame,b^byte(i%4+1))}
 return frame
}
func TestFragmentedMessageWithPing(t *testing.T){
 data:=append(masked(opcodeText,false,[]byte("hel")),masked(opcodePing,true,[]byte("p"))...)
 data=append(data,masked(opcodeContinuation,true,[]byte("lo"))...)
 var response bytes.Buffer
 got,err:=ReadMessage(bytes.NewReader(data),&response)
 if err!=nil||string(got)!="hello"{t.Fatalf("message %q, %v",got,err)}
 if !bytes.Equal(response.Bytes(),[]byte{0x8a,1,'p'}){t.Fatalf("pong %x",response.Bytes())}
}
func TestClose(t *testing.T){
 var response bytes.Buffer
 _,err:=ReadMessage(bytes.NewReader(masked(opcodeClose,true,nil)),&response)
 if !errors.Is(err,ErrClosed){t.Fatalf("close: %v",err)}
 if !bytes.Equal(response.Bytes(),[]byte{0x88,0}){t.Fatalf("close response: %x",response.Bytes())}
}
func TestInvalidFrames(t *testing.T){
 cases:=[][]byte{
  {0x81,0x01,'x'}, // unmasked
  masked(opcodeContinuation,true,[]byte("x")), // unexpected continuation
  masked(2,true,[]byte("x")), // binary
  masked(opcodeText,true,[]byte{0xff}), // invalid utf8
  masked(opcodePing,false,[]byte("x")), // fragmented control
  {0x81,0xfe,0x10,0x01}, // oversized
 }
 for i,data:=range cases{
  _,err:=ReadMessage(bytes.NewReader(data),&bytes.Buffer{})
  if !errors.Is(err,ErrFrame){t.Errorf("case %d: %v",i,err)}
 }
}
