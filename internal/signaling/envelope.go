// Package signaling defines validation for future authenticated signaling.
// No network transport or client authentication is implemented here.
package signaling

import (
 "errors"
 "strings"
)

const MaxSDPBytes = 65536

type Type string
const (
 Offer Type = "offer"
 Answer Type = "answer"
 Candidate Type = "candidate"
)

type Envelope struct {
 Version int `json:"version"`
 SessionID string `json:"session_id"`
 SenderID string `json:"sender_id"`
 RecipientID string `json:"recipient_id"`
 Type Type `json:"type"`
 Payload string `json:"payload"`
}

func (e Envelope) Validate()error{
 if e.Version!=1{return errors.New("unsupported signaling version")}
 if e.SessionID==""||e.SenderID==""||e.RecipientID==""{return errors.New("missing routing identity")}
 if e.SenderID==e.RecipientID{return errors.New("self-addressed signaling not allowed")}
 switch e.Type{case Offer,Answer,Candidate:default:return errors.New("unsupported signaling type")}
 if len(e.Payload)==0||len(e.Payload)>MaxSDPBytes{return errors.New("invalid payload size")}
 if strings.ContainsRune(e.Payload,0){return errors.New("NUL byte in payload")}
 return nil
}
