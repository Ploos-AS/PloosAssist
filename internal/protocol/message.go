// Package protocol defines the versioned signaling message envelope.
package protocol

import (
 "encoding/json"
 "errors"
)

const Version = 1

type Kind string

const (
 KindHello Kind = "hello"
 KindError Kind = "error"
)

// Message is a versioned envelope. Unrecognized kinds must be rejected by consumers.
type Message struct {
 Version int `json:"version"`
 Kind Kind `json:"kind"`
 SessionID string `json:"session_id,omitempty"`
 Payload json.RawMessage `json:"payload,omitempty"`
}

func (m Message) Validate() error {
 if m.Version != Version { return errors.New("unsupported protocol version") }
 switch m.Kind { case KindHello, KindError: default: return errors.New("unsupported message kind") }
 if len(m.Payload)>65536 { return errors.New("payload too large") }
 return nil
}
