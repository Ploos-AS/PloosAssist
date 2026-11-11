// Package identity provides ephemeral Ed25519 proof-of-possession primitives.
// This package does not authenticate a remote network connection by itself.
package identity

import (
 "crypto/ed25519"
 "crypto/rand"
 "crypto/sha256"
 "encoding/base64"
 "errors"
 "time"
)

const ChallengeSize = 32

var (
 ErrInvalidKey = errors.New("invalid public key")
 ErrInvalidSignature = errors.New("invalid signature")
 ErrExpired = errors.New("challenge expired")
)

type KeyPair struct {
 Public ed25519.PublicKey
 Private ed25519.PrivateKey
}

func Generate() (KeyPair,error) {
 public,private,err:=ed25519.GenerateKey(rand.Reader)
 if err!=nil{return KeyPair{},err}
 return KeyPair{Public:public,Private:private},nil
}

func Fingerprint(public ed25519.PublicKey)(string,error){
 if len(public)!=ed25519.PublicKeySize{return "",ErrInvalidKey}
 digest:=sha256.Sum256(public)
 return base64.RawURLEncoding.EncodeToString(digest[:]),nil
}

type Challenge struct {
 Nonce []byte
 ExpiresAt time.Time
}

// NewChallenge produces a fresh nonce for a single authentication attempt.
// The caller MUST store it server-side, bind it to a connection, and consume it once.
func NewChallenge(now time.Time,ttl time.Duration)(Challenge,error){
 if ttl<=0 || ttl>time.Minute{return Challenge{},errors.New("challenge ttl out of range")}
 nonce:=make([]byte,ChallengeSize)
 if _,err:=rand.Read(nonce);err!=nil{return Challenge{},err}
 return Challenge{Nonce:nonce,ExpiresAt:now.Add(ttl)},nil
}

func transcript(nonce []byte)[]byte{
 prefix:=[]byte("PloosAssist/identity/v1\x00")
 out:=make([]byte,0,len(prefix)+len(nonce))
 out=append(out,prefix...)
 return append(out,nonce...)
}

func Sign(private ed25519.PrivateKey,challenge Challenge)([]byte,error){
 if len(private)!=ed25519.PrivateKeySize || len(challenge.Nonce)!=ChallengeSize{return nil,errors.New("invalid signing inputs")}
 return ed25519.Sign(private,transcript(challenge.Nonce)),nil
}

// Verify checks cryptographic possession and expiry only.
// It does not provide replay protection, peer trust, or session authorization.
func Verify(public ed25519.PublicKey,challenge Challenge,signature []byte,now time.Time)error{
 if len(public)!=ed25519.PublicKeySize || len(challenge.Nonce)!=ChallengeSize{return ErrInvalidKey}
 if !now.Before(challenge.ExpiresAt){return ErrExpired}
 if len(signature)!=ed25519.SignatureSize || !ed25519.Verify(public,transcript(challenge.Nonce),signature){return ErrInvalidSignature}
 return nil
}
