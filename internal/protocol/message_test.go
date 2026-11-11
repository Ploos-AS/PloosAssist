package protocol

import "testing"

func TestValidate(t *testing.T) {
 tests := []struct{name string; msg Message; ok bool}{
  {"valid",Message{Version:Version,Kind:KindHello},true},
  {"version",Message{Version:99,Kind:KindHello},false},
  {"kind",Message{Version:Version,Kind:"remote-input"},false},
 }
 for _,tt:=range tests { t.Run(tt.name,func(t *testing.T){err:=tt.msg.Validate();if (err==nil)!=tt.ok {t.Fatalf("Validate() error = %v, want ok=%v",err,tt.ok)}}) }
}
