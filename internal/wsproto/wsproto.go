// Package wsproto implements a deliberately limited RFC 6455 development transport.
// It supports masked, unfragmented, small text frames only. It is NOT a production WebSocket stack.
package wsproto

import (
 "bufio"
 "crypto/sha1"
 "encoding/base64"
 "encoding/binary"
 "errors"
 "io"
 "net"
 "net/http"
 "strings"
)

const MaxFrame = 4096
var ErrFrame = errors.New("invalid websocket frame")

func Upgrade(w http.ResponseWriter,r *http.Request)(net.Conn,*bufio.ReadWriter,error){
 if r.Method!="GET"||!strings.EqualFold(r.Header.Get("Upgrade"),"websocket")||!strings.Contains(strings.ToLower(r.Header.Get("Connection")),"upgrade")||r.Header.Get("Sec-WebSocket-Version")!="13"{
  return nil,nil,ErrFrame
 }
 key:=r.Header.Get("Sec-WebSocket-Key")
 decoded,err:=base64.StdEncoding.DecodeString(key)
 if err!=nil||len(decoded)!=16{return nil,nil,ErrFrame}
 hijacker,ok:=w.(http.Hijacker);if !ok{return nil,nil,ErrFrame}
 conn,rw,err:=hijacker.Hijack();if err!=nil{return nil,nil,err}
 digest:=sha1.Sum([]byte(key+"258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
 _,err=rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "+base64.StdEncoding.EncodeToString(digest[:])+"\r\n\r\n")
 if err==nil{err=rw.Flush()}
 if err!=nil{conn.Close();return nil,nil,err}
 return conn,rw,nil
}

func ReadText(r io.Reader)([]byte,error){
 var header [2]byte
 if _,err:=io.ReadFull(r,header[:]);err!=nil{return nil,err}
 // Only final, masked text frames; reject RSV bits and all control/continuation frames.
 if header[0]!=0x81 || header[1]&0x80==0{return nil,ErrFrame}
 size:=int(header[1]&0x7f)
 if size==126{
  var extended [2]byte
  if _,err:=io.ReadFull(r,extended[:]);err!=nil{return nil,err}
  size=int(binary.BigEndian.Uint16(extended[:]))
 }else if size==127{return nil,ErrFrame}
 if size>MaxFrame{return nil,ErrFrame}
 var mask [4]byte
 if _,err:=io.ReadFull(r,mask[:]);err!=nil{return nil,err}
 payload:=make([]byte,size)
 if _,err:=io.ReadFull(r,payload);err!=nil{return nil,err}
 for i:=range payload{payload[i]^=mask[i%4]}
 return payload,nil
}
func WriteText(w io.Writer,data []byte)error{
 if len(data)>MaxFrame{return ErrFrame}
 header:=[]byte{0x81,byte(len(data))}
 if len(data)>=126{header=[]byte{0x81,126,byte(len(data)>>8),byte(len(data))}}
 _,err:=w.Write(append(header,data...))
 return err
}
