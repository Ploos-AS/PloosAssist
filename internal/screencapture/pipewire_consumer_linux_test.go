//go:build linux

package screencapture

import (
 "context"
 "errors"
 "testing"

 "github.com/godbus/dbus/v5"
)

func TestPortalStreamValidation(t *testing.T){
 rows:=[]struct{NodeID uint32;Properties map[string]dbus.Variant}{
  {NodeID:42,Properties:map[string]dbus.Variant{}},
 }
 streams,err:=parseStreams(rows)
 if err!=nil||len(streams)!=1||streams[0].NodeID!=42{t.Fatalf("valid streams: %v %v",streams,err)}
 for _,bad:=range []any{
  nil,
  []struct{NodeID uint32;Properties map[string]dbus.Variant}{},
  []struct{NodeID uint32;Properties map[string]dbus.Variant}{{NodeID:0}},
  []struct{NodeID uint32;Properties map[string]dbus.Variant}{{NodeID:1},{NodeID:1}},
 }{
  if _,err:=parseStreams(bad);err==nil{t.Fatalf("accepted invalid stream %#v",bad)}
 }
}
func TestPortalStreamAccessRequiresStart(t *testing.T){
 s:=&dbusPortalSession{}
 if _,err:=s.Streams();!errors.Is(err,ErrNotAuthorized){t.Fatalf("before Start: %v",err)}
 s.started=true
 s.streams=[]Stream{{NodeID:42}}
 got,err:=s.Streams()
 if err!=nil||got[0].NodeID!=42{t.Fatalf("after Start: %v %v",got,err)}
 got[0].NodeID=99
 again,_:=s.Streams()
 if again[0].NodeID!=42{t.Fatal("returned mutable backing data")}
 s.closed=true
 if _,err:=s.Streams();!errors.Is(err,ErrStopped){t.Fatalf("after Close: %v",err)}
}
func TestLocalCaptureFailsClosedWithoutNativeConsumer(t *testing.T){
 c:=NewLocalCapture()
 if err:=c.Run(context.Background(),nil,nil);!errors.Is(err,ErrNotAuthorized){t.Fatal(err)}
 c.Stop()
 if _,ok:=c.Snapshot();ok{t.Fatal("frame after Stop")}
}
