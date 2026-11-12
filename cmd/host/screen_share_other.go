//go:build !linux

package main

import (
 "context"
 "errors"
)

type screenShareAction struct{}
func newScreenShareAction()*screenShareAction{return &screenShareAction{}}
func(*screenShareAction)Start(context.Context)error{return errors.New("screen sharing requires Linux")}
func(*screenShareAction)Stop(){}
func(*screenShareAction)Status()string{return "unavailable"}
