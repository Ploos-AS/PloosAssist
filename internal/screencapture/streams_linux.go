//go:build linux

package screencapture

import (
 "fmt"

 "github.com/godbus/dbus/v5"
)

// Stream identifies one node selected by the desktop portal.
// This is metadata only; it is not a captured frame.
type Stream struct{NodeID uint32}
func parseStreams(value any)([]Stream,error){
 // D-Bus signature a(ua{sv}) is decoded as a slice of structs by
 // godbus. Accept only the exact shape and nonzero node identifiers.
 rows,ok:=value.([]struct{
  NodeID uint32
  Properties map[string]dbus.Variant
 })
 if !ok||len(rows)==0||len(rows)>8{return nil,ErrPortalResponse}
 streams:=make([]Stream,0,len(rows))
 seen:=make(map[uint32]bool,len(rows))
 for _,row:=range rows{
  if row.NodeID==0||seen[row.NodeID]{return nil,fmt.Errorf("%w: invalid stream node",ErrPortalResponse)}
  seen[row.NodeID]=true
  streams=append(streams,Stream{NodeID:row.NodeID})
 }
 return streams,nil
}
