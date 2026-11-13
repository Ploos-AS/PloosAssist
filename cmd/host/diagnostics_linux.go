//go:build linux

package main

import (
 "fmt"
 "os"
 "os/exec"
 "runtime"
 "strings"

 "github.com/Ploos-AS/PloosAssist/internal/screencapture"
)

// printCaptureDiagnostics reports prerequisites only. It never starts a
// portal session, requests screen permission or captures any pixels.
func printCaptureDiagnostics(){
 env:=screencapture.Probe()
 fmt.Println("PloosAssist Linux capture diagnostics (no screen capture)")
 fmt.Println("OS:",runtime.GOOS,"architecture:",runtime.GOARCH)
 fmt.Println("XDG_SESSION_TYPE:",safeDiagnosticValue(env.SessionType))
 fmt.Println("XDG_CURRENT_DESKTOP:",safeDiagnosticValue(os.Getenv("XDG_CURRENT_DESKTOP")))
 fmt.Println("WAYLAND_DISPLAY:",present(os.Getenv("WAYLAND_DISPLAY")))
 fmt.Println("DBUS_SESSION_BUS_ADDRESS:",present(os.Getenv("DBUS_SESSION_BUS_ADDRESS")))
 fmt.Println("xdg-desktop-portal executable:",env.PortalAvailable)
 fmt.Println("pipewire executable:",env.PipeWireAvailable)
 fmt.Println("native PipeWire backend compiled:",localVideoConsumer()!=nil)
 for _,name:=range []string{"pipewire","wireplumber","xdg-desktop-portal"}{
  path,err:=exec.LookPath(name)
  if err!=nil{fmt.Println(name+": not found");continue}
  fmt.Println(name+":",path)
 }
 if env.SessionType!="wayland"||!env.PortalAvailable||!env.PipeWireAvailable{
  fmt.Println("RESULT: prerequisites incomplete; no capture attempted")
 }else if localVideoConsumer()==nil{
  fmt.Println("RESULT: prerequisites found; rebuild with -tags pipewire for native capture")
 }else{
  fmt.Println("RESULT: prerequisites found; actual desktop portal approval and frame delivery NOT tested")
 }
}
func safeDiagnosticValue(value string)string{
 value=strings.TrimSpace(value)
 if value==""{return "(unset)"}
 if len(value)>120{value=value[:120]}
 return strings.Map(func(r rune)rune{
  if r<32||r==127{return -1}
  return r
 },value)
}
func present(value string)string{
 if value==""{return "unset"}
 return "set (value hidden)"
}
