// Command host is an experimental Linux host-side connection diagnostic.
package main

import (
 "context"
 "flag"
 "fmt"
 "log"
 "net"
 "net/http"
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
 hostCapability:=flag.String("host-capability","","development host capability (enables local consent API)")
 uiAddress:=flag.String("ui-listen","127.0.0.1:0","local consent API listen address")
 flag.Parse()
 if *session==""{log.Fatal("-session is required")}
 key,err:=identity.Generate()
 if err!=nil{log.Fatal(err)}
 ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM)
 defer stop()
 conn,fingerprint,err:=hostclient.Connect(ctx,*endpoint,*session,key)
 if err!=nil{log.Fatal(err)}
 defer conn.Close()
 fmt.Println("PloosAssist host authenticated. Fingerprint:",fingerprint)
 if *hostCapability==""{
  fmt.Println("Identity-only diagnostic. No consent UI, screen sharing or remote control. Ctrl+C to disconnect.")
  done:=make(chan error,1)
  go func(){for{if _,_,err:=conn.ReadMessage();err!=nil{done<-err;return}}}()
  select{
  case <-ctx.Done():
   _=conn.WriteControl(websocket.CloseMessage,websocket.FormatCloseMessage(websocket.CloseNormalClosure,""),time.Now().Add(time.Second))
  case err:=<-done:log.Printf("connection ended: %v",err)
  }
  return
 }
 listener,err:=net.Listen("tcp",*uiAddress)
 if err!=nil{log.Fatal(err)}
 defer listener.Close()
 if listener.Addr().(*net.TCPAddr).IP==nil||!listener.Addr().(*net.TCPAddr).IP.IsLoopback(){log.Fatal("consent API must bind loopback")}
 origin:="http://127.0.0.1:"+fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
 client:=hostclient.NewConsentClient(conn)
 if err:=client.EnrollHost(*hostCapability);err!=nil{log.Fatal("host enrollment failed: ",err)}
 bridge,err:=hostclient.NewConsentBridge(origin,*session,client.Decide)
 if err!=nil{log.Fatal(err)}
 server:=&http.Server{Handler:bridge.Handler(),ReadHeaderTimeout:5*time.Second}
 go func(){if err:=server.Serve(listener);err!=nil&&err!=http.ErrServerClosed{log.Printf("local bridge stopped: %v",err);stop()}}()
 defer func(){shutdown,cancel:=context.WithTimeout(context.Background(),3*time.Second);defer cancel();_=server.Shutdown(shutdown)}()
 fmt.Println("Local consent API initialized at",origin,"(no browser UI served yet)")
 fmt.Println("Token is intentionally not printed. Remote control remains disabled.")
 // The same goroutine owns all websocket reads and writes; HTTP decision
 // callbacks serialize through this goroutine to avoid concurrent reads.
 type decision struct{ticket string;approve bool;reply chan error}
 decisions:=make(chan decision)
 // Override bridge callback using a serialized channel in the next revision;
 // until then, do not poll concurrently with HTTP decisions.
 _=decisions
 ticker:=time.NewTicker(2*time.Second)
 defer ticker.Stop()
 for{
  select{
  case <-ctx.Done():bridge.Clear();return
  case <-ticker.C:
   pending,ok,err:=client.Check()
   if err!=nil{bridge.Clear();log.Printf("host connection ended: %v",err);return}
   if ok{bridge.SetPending(pending)}else{bridge.Clear()}
  }
 }
}
