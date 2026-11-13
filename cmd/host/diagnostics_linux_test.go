//go:build linux

package main

import (
 "bytes"
 "os"
 "strings"
 "testing"
)

func TestDiagnosticValuesHideSessionBusDetails(t *testing.T){
 if got:=present("unix:path=/run/user/1000/bus");got!="set (value hidden)"{
  t.Fatalf("bus address leaked: %q",got)
 }
 if got:=present("");got!="unset"{t.Fatalf("empty bus address: %q",got)}
 if got:=safeDiagnosticValue("GNOME\nsecret");strings.Contains(got,"\n")||strings.Contains(got,"secret")==false{
  t.Fatalf("diagnostic value: %q",got)
 }
}
func TestDiagnosticsDoesNotRequireSessionOrCapture(t *testing.T){
 old:=os.Stdout
 r,w,err:=os.Pipe()
 if err!=nil{t.Fatal(err)}
 os.Stdout=w
 defer func(){os.Stdout=old}()
 printCaptureDiagnostics()
 _=w.Close()
 var buf bytes.Buffer
 _,err=buf.ReadFrom(r)
 _=r.Close()
 if err!=nil{t.Fatal(err)}
 output:=buf.String()
 if !strings.Contains(output,"no screen capture")||!strings.Contains(output,"actual desktop portal approval and frame delivery NOT tested") &&
 !strings.Contains(output,"no capture attempted") && !strings.Contains(output,"rebuild with -tags pipewire"){
  t.Fatalf("diagnostic output must not claim screen capture: %q",output)
 }
}
