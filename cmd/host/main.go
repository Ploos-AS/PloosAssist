// Command host is an experimental Linux host-side connection diagnostic.
package main

import (
 "context"
 "flag"
 "fmt"
 "log"
 "os"
 "os/signal"
 "syscall"
 "time"

 "github.com/Ploos-AS/PloosAssist/internal/hostclient"
 "github.com/Ploos-AS/PloosAssist/internal/identity"
 "github.com/gorilla/websocket"
)

func main(){
 endpoint:=flag.String("server","ws://127.0.0.1:8080/v1/ws","loopback development server")
 session:=flag.String("session","","development session identifier (required)")
 flag.Parse()
 if *session==""{log.Fatal("-session is required")}
 key,err:=identity.Generate()
 if err!=nil{log.Fatal(err)}
 ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
 defer stop()
 conn,fingerprint,err:=hostclient.Connect(ctx,*endpoint,*session,key)
 if err!=nil{log.Fatal(err)}
 defer conn.Close()
 fmt.Println("PloosAssist host diagnostic authenticated. Fingerprint:",fingerprint)
 fmt.Println("No approval requests, screen sharing or remote control are supported. Ctrl+C to disconnect.")
 done:=make(chan error,1)
 go func(){
  for {
   if _,_,err:=conn.ReadMessage();err!=nil{done<-err;return}
  }
 }()
 select{
 case <-ctx.Done():
  _=conn.WriteControl(websocket.CloseMessage,websocket.FormatCloseMessage(websocket.CloseNormalClosure,""),time.Now().Add(time.Second))
 case err:=<-done:
  log.Printf("connection ended: %v",err)
 }
}
