// Package hostclient implements the local, identity-only WebSocket handshake.
package hostclient

import (
 "context"
 "crypto/ed25519"
 "encoding/base64"
 "encoding/json"
 "errors"
 "fmt"
 "net"
 "net/url"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/gorilla/websocket"
)

var ErrProtocol=errors.New("invalid host handshake response")

// Connect authenticates an ephemeral key to the development server.
// It never authorizes remote control or requests screen sharing.
func Connect(ctx context.Context, endpoint,session string,key identity.KeyPair)(*websocket.Conn,string,error){
 if len(key.Public)!=ed25519.PublicKeySize||len(key.Private)!=ed25519.PrivateKeySize||session==""||len(session)>128{return nil,"",ErrProtocol}
 u,err:=url.Parse(endpoint)
 if err!=nil{return nil,"",err}
 if u.Scheme!="ws"||u.Host==""||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Path!="/v1/ws"{return nil,"",errors.New("only local ws://.../v1/ws supported")}
 host,_,err:=net.SplitHostPort(u.Host)
 if err!=nil{return nil,"",err}
 ip:=net.ParseIP(host)
 if ip==nil||!ip.IsLoopback(){return nil,"",errors.New("remote connections are disabled")}
 q:=u.Query()
 q.Set("session_id",session)
 q.Set("public_key",base64.RawURLEncoding.EncodeToString(key.Public))
 u.RawQuery=q.Encode()
 dialer:=websocket.Dialer{HandshakeTimeout:5*time.Second,Proxy:nil}
 conn,_,err:=dialer.DialContext(ctx,u.String(),nil)
 if err!=nil{return nil,"",err}
 fail:=func(err error)(*websocket.Conn,string,error){_ =conn.Close();return nil,"",err}
 conn.SetReadLimit(4096)
 _=conn.SetReadDeadline(time.Now().Add(15*time.Second))
 kind,data,err:=conn.ReadMessage()
 if err!=nil{return fail(err)}
 if kind!=websocket.TextMessage{return fail(ErrProtocol)}
 var challenge struct{Type string `json:"type"`;Nonce string `json:"nonce"`;ID string `json:"challenge_id"`}
 if err:=json.Unmarshal(data,&challenge);err!=nil{return fail(err)}
 if challenge.Type!="challenge"||challenge.ID==""{return fail(ErrProtocol)}
 nonce,err:=base64.RawURLEncoding.DecodeString(challenge.Nonce)
 if err!=nil||len(nonce)!=identity.ChallengeSize{return fail(ErrProtocol)}
 sig,err:=identity.Sign(key.Private,identity.Challenge{Nonce:nonce})
 if err!=nil{return fail(err)}
 proof,_:=json.Marshal(map[string]string{"type":"proof","signature":base64.RawURLEncoding.EncodeToString(sig)})
 if err:=conn.WriteMessage(websocket.TextMessage,proof);err!=nil{return fail(err)}
 kind,data,err=conn.ReadMessage()
 if err!=nil{return fail(err)}
 if kind!=websocket.TextMessage{return fail(ErrProtocol)}
 var response struct{Type string `json:"type"`;Fingerprint string `json:"fingerprint"`}
 if err:=json.Unmarshal(data,&response);err!=nil{return fail(err)}
 expected,err:=identity.Fingerprint(key.Public)
 if err!=nil{return fail(err)}
 if response.Type!="proof_verified"||response.Fingerprint!=expected{return fail(fmt.Errorf("%w: fingerprint mismatch",ErrProtocol))}
 _=conn.SetReadDeadline(time.Time{})
 return conn,expected,nil
}
