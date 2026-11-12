//go:build linux

package main

import "testing"

func TestScreenShareStopIsFinal(t *testing.T){
 a:=newScreenShareAction()
 if got:=a.Status();got!="idle"{t.Fatalf("initial status %q",got)}
 a.Stop()
 if got:=a.Status();got!="stopped"{t.Fatalf("stop status %q",got)}
 a.setOutcome(nil,true)
 if got:=a.Status();got!="stopped"{t.Fatalf("late callback changed stop status: %q",got)}
 a.Stop()
 if got:=a.Status();got!="stopped"{t.Fatalf("repeated stop status %q",got)}
}
