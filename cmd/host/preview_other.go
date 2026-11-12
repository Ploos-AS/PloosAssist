//go:build !linux

package main

import "errors"

func(a *screenShareAction)PreviewPNG()([]byte,error){
 return nil,errors.New("local preview unsupported")
}
