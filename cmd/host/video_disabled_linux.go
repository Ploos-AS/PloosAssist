//go:build linux && (!cgo || !pipewire)

package main

import "github.com/Ploos-AS/PloosAssist/internal/screencapture"

func localVideoConsumer()screencapture.VideoConsumer{return nil}
