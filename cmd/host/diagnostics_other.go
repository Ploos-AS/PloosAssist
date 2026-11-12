//go:build !linux

package main

import "fmt"

func printCaptureDiagnostics(){
 fmt.Println("Linux screen capture diagnostics are unavailable on this operating system.")
}
