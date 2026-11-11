package identity

import (
 "crypto/ed25519"
 "crypto/rand"
 "encoding/base64"
 "errors"
 "sync"
 "time"
)

var (
 ErrChallengeNotFound = errors.New("challenge unavailable")
 ErrChallengeContext = errors.New("challenge context mismatch")
 ErrChallengeCapacity = errors.New("challenge capacity reached")
)

type challengeEntry struct {
 challenge Challenge
 peer ed25519.PublicKey
 connectionID string
 sessionID string
}

// ChallengeStore tracks single-use proof-of-possession challenges.
// It is in-memory and does not authenticate the connection identifier itself.
type ChallengeStore struct {
 mu sync.Mutex
 entries map[string]challengeEntry
 now func() time.Time
 ttl time.Duration
 max int
}

func NewChallengeStore(ttl time.Duration,max int)(*ChallengeStore,error){
 if ttl<=0 || ttl>time.Minute || max<=0{return nil,errors.New("invalid challenge store settings")}
 return &ChallengeStore{entries:make(map[string]challengeEntry),now:time.Now,ttl:ttl,max:max},nil
}

// Issue binds a nonce to a peer key and externally assigned connection/session IDs.
func (s *ChallengeStore) Issue(peer ed25519.PublicKey,connectionID,sessionID string)(string,Challenge,error){
 s.mu.Lock();defer s.mu.Unlock()
 if len(peer)!=ed25519.PublicKeySize || connectionID=="" || sessionID=="" {return "",Challenge{},errors.New("invalid challenge context")}
 for id,e:=range s.entries {if !s.now().Before(e.challenge.ExpiresAt){delete(s.entries,id)}}
 if len(s.entries)>=s.max{return "",Challenge{},ErrChallengeCapacity}
 challenge,err:=NewChallenge(s.now(),s.ttl);if err!=nil{return "",Challenge{},err}
 token:=make([]byte,32);if _,err=rand.Read(token);err!=nil{return "",Challenge{},err}
 id:=base64.RawURLEncoding.EncodeToString(token)
 s.entries[id]=challengeEntry{challenge:challenge,peer:append(ed25519.PublicKey(nil),peer...),connectionID:connectionID,sessionID:sessionID}
 return id,challenge,nil
}

// Consume atomically removes the challenge BEFORE verifying a proof.
// Even a failed attempt consumes it; replaying a successful proof cannot work.
func (s *ChallengeStore) Consume(id string,peer ed25519.PublicKey,connectionID,sessionID string,signature []byte)error{
 s.mu.Lock()
 entry,ok:=s.entries[id]
 if ok{delete(s.entries,id)}
 s.mu.Unlock()
 if !ok{return ErrChallengeNotFound}
 if connectionID!=entry.connectionID || sessionID!=entry.sessionID || len(peer)!=ed25519.PublicKeySize || !peer.Equal(entry.peer){
  return ErrChallengeContext
 }
 return Verify(entry.peer,entry.challenge,signature,s.now())
}
