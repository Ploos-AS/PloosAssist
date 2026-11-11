package wsproto

import (
 "encoding/binary"
 "errors"
 "io"
 "unicode/utf8"
)

var ErrClosed = errors.New("websocket peer closed")
const (
 opcodeContinuation = 0
 opcodeText = 1
 opcodeClose = 8
 opcodePing = 9
 opcodePong = 10
)

// ReadMessage reads one bounded text message, responding to ping frames.
// It supports fragmented text frames and rejects binary and invalid UTF-8.
// Only one goroutine may read or write a given stream at a time.
func ReadMessage(r io.Reader,w io.Writer)([]byte,error){
 var message []byte
 fragmented:=false
 for {
  var header [2]byte
  if _,err:=io.ReadFull(r,header[:]);err!=nil{return nil,err}
  if header[0]&0x70!=0 || header[1]&0x80==0{return nil,ErrFrame}
  final:=header[0]&0x80!=0
  opcode:=header[0]&0x0f
  size:=uint64(header[1]&0x7f)
  if size==126{
   var ext [2]byte
   if _,err:=io.ReadFull(r,ext[:]);err!=nil{return nil,err}
   size=uint64(binary.BigEndian.Uint16(ext[:]))
   if size<126{return nil,ErrFrame}
  }else if size==127{
   var ext [8]byte
   if _,err:=io.ReadFull(r,ext[:]);err!=nil{return nil,err}
   size=binary.BigEndian.Uint64(ext[:])
   if size<65536 || size>>63!=0{return nil,ErrFrame}
  }
  control:=opcode>=8
  if control && (!final || size>125){return nil,ErrFrame}
  if size>MaxFrame || uint64(len(message))+size>MaxFrame{return nil,ErrFrame}
  var mask [4]byte
  if _,err:=io.ReadFull(r,mask[:]);err!=nil{return nil,err}
  data:=make([]byte,int(size))
  if _,err:=io.ReadFull(r,data);err!=nil{return nil,err}
  for i:=range data{data[i]^=mask[i%4]}
  switch opcode{
  case opcodePing:
   if err:=writeControl(w,opcodePong,data);err!=nil{return nil,err}
  case opcodePong:
   // Ignore unsolicited pong.
  case opcodeClose:
   if len(data)==1{return nil,ErrFrame}
   _=writeControl(w,opcodeClose,data)
   return nil,ErrClosed
  case opcodeText:
   if fragmented{return nil,ErrFrame}
   message=append(message,data...)
   if final{
    if !utf8.Valid(message){return nil,ErrFrame}
    return message,nil
   }
   fragmented=true
  case opcodeContinuation:
   if !fragmented{return nil,ErrFrame}
   message=append(message,data...)
   if final{
    if !utf8.Valid(message){return nil,ErrFrame}
    return message,nil
   }
  default:
   return nil,ErrFrame
  }
 }
}

func writeControl(w io.Writer,opcode byte,data []byte)error{
 if len(data)>125{return ErrFrame}
 frame:=append([]byte{0x80|opcode,byte(len(data))},data...)
 _,err:=w.Write(frame)
 return err
}
