// Package session implements an in-memory consent-first pairing state machine.
// It does not authenticate network peers or grant remote-control permissions.
package session

import (
 "crypto/rand"
 "crypto/subtle"
 "encoding/base64"
 "errors"
 "sync"
 "time"
)

var (
 ErrNotFound = errors.New("session not found")
 ErrExpired = errors.New("session expired")
 ErrInvalidState = errors.New("invalid session state")
 ErrUnauthorized = errors.New("invalid capability")
)

type State string
const (
 Pending State = "pending"
 AwaitingApproval State = "awaiting_approval"
 Approved State = "approved"
 Denied State = "denied"
 Revoked State = "revoked"
)

type Pair struct {
 ID string
 State State
 ExpiresAt time.Time
 hostSecret string
 technicianSecret string
}

type Store struct {
 mu sync.Mutex
 pairs map[string]*Pair
 now func() time.Time
 ttl time.Duration
}

func NewStore(ttl time.Duration) (*Store,error) {
 if ttl<=0 || ttl>15*time.Minute {return nil,errors.New("ttl must be between 1ns and 15m")}
 return &Store{pairs:make(map[string]*Pair),now:time.Now,ttl:ttl},nil
}
func secret() (string,error) {
 b:=make([]byte,32)
 if _,err:=rand.Read(b);err!=nil{return "",err}
 return base64.RawURLEncoding.EncodeToString(b),nil
}
func (s *Store) Create() (id,hostSecret string,err error) {
 s.mu.Lock();defer s.mu.Unlock()
 for {
  id,err=secret();if err!=nil{return "","",err}
  if _,exists:=s.pairs[id];!exists{break}
 }
 hostSecret,err=secret();if err!=nil{return "","",err}
 s.pairs[id]=&Pair{ID:id,State:Pending,ExpiresAt:s.now().Add(s.ttl),hostSecret:hostSecret}
 return id,hostSecret,nil
}
func (s *Store) lookup(id string)(*Pair,error){
 p,ok:=s.pairs[id];if !ok{return nil,ErrNotFound}
 if !s.now().Before(p.ExpiresAt) {delete(s.pairs,id);return nil,ErrExpired}
 return p,nil
}
// Request creates a technician capability; it never approves the session.
func (s *Store) Request(id string)(string,error){
 s.mu.Lock();defer s.mu.Unlock()
 p,err:=s.lookup(id);if err!=nil{return "",err}
 if p.State!=Pending{return "",ErrInvalidState}
 cap,err:=secret();if err!=nil{return "",err}
 p.technicianSecret=cap;p.State=AwaitingApproval
 return cap,nil
}
// Decide requires the host's independent capability and explicit approval.
func (s *Store) Decide(id,hostSecret string,approve bool) error {
 s.mu.Lock();defer s.mu.Unlock()
 p,err:=s.lookup(id);if err!=nil{return err}
 if subtle.ConstantTimeCompare([]byte(p.hostSecret),[]byte(hostSecret))!=1{return ErrUnauthorized}
 if p.State!=AwaitingApproval{return ErrInvalidState}
 if approve {p.State=Approved}else{p.State=Denied}
 return nil
}
func (s *Store) Revoke(id,hostSecret string)error{
 s.mu.Lock();defer s.mu.Unlock()
 p,err:=s.lookup(id);if err!=nil{return err}
 if p.hostSecret!=hostSecret{return ErrUnauthorized}
 if p.State==Revoked || p.State==Denied{return ErrInvalidState}
 p.State=Revoked
 return nil
}
// Authorized only indicates pairing approval; it must not be used as a
// substitute for authenticated transport and per-action remote-input consent.
func (s *Store) Authorized(id,technicianSecret string)bool{
 s.mu.Lock();defer s.mu.Unlock()
 p,err:=s.lookup(id)
 return err==nil && p.State==Approved && p.technicianSecret!="" && subtle.ConstantTimeCompare([]byte(p.technicianSecret),[]byte(technicianSecret))==1
}

// VerifyRoleCapability checks a pairing role credential without changing the
// pairing state. It returns the original pairing expiry so downstream role
// bindings cannot outlive the pairing. Callers must bind the credential to a
// verified connection, never trust a client-supplied connection ID.
func (s *Store) VerifyRoleCapability(id,role,capability string)(time.Time,error){
 s.mu.Lock()
 defer s.mu.Unlock()
 p,err:=s.lookup(id)
 if err!=nil{return time.Time{},err}
 if capability==""{return time.Time{},ErrUnauthorized}
 var expected string
 switch role {
 case "host":
  if p.State!=Pending&&p.State!=AwaitingApproval&&p.State!=Approved{return time.Time{},ErrInvalidState}
  expected=p.hostSecret
 case "technician":
  if p.State!=AwaitingApproval&&p.State!=Approved{return time.Time{},ErrInvalidState}
  expected=p.technicianSecret
 default:return time.Time{},ErrUnauthorized
 }
 if expected==""||subtle.ConstantTimeCompare([]byte(expected),[]byte(capability))!=1{return time.Time{},ErrUnauthorized}
 return p.ExpiresAt,nil
}
