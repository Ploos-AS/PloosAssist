package signaling

import (
 "strings"
 "testing"
)

func TestValidate(t *testing.T){
 good:=Envelope{Version:1,SessionID:"session",SenderID:"host",RecipientID:"tech",Type:Offer,Payload:"sdp"}
 if err:=good.Validate();err!=nil{t.Fatal(err)}
 tests:=[]struct{name string;change func(*Envelope)}{
  {"version",func(e *Envelope){e.Version=2}},
  {"sender",func(e *Envelope){e.SenderID=""}},
  {"self",func(e *Envelope){e.RecipientID=e.SenderID}},
  {"kind",func(e *Envelope){e.Type="remote-input"}},
  {"empty",func(e *Envelope){e.Payload=""}},
  {"oversize",func(e *Envelope){e.Payload=strings.Repeat("x",MaxSDPBytes+1)}},
  {"nul",func(e *Envelope){e.Payload="x\x00y"}},
 }
 for _,tt:=range tests{t.Run(tt.name,func(t *testing.T){v:=good;tt.change(&v);if err:=v.Validate();err==nil{t.Fatal("expected validation error")}})}
}
