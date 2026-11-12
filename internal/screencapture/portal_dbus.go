//go:build linux

package screencapture

import (
 "context"
 "crypto/rand"
 "encoding/hex"
 "errors"
 "fmt"
 "os"
 "sync"
 "time"

 "github.com/godbus/dbus/v5"
)

const (
 portalBus="org.freedesktop.portal.Desktop"
 portalPath=dbus.ObjectPath("/org/freedesktop/portal/desktop")
 portalScreenCast="org.freedesktop.portal.ScreenCast"
 portalRequest="org.freedesktop.portal.Request"
 portalSession="org.freedesktop.portal.Session"
)
var ErrPortalResponse=errors.New("desktop portal denied or failed the request")

// DBusPortal uses the user's session bus. A successful ScreenCast Start
// establishes a portal permission grant; it does not read or send frames.
type DBusPortal struct{conn *dbus.Conn}
func NewDBusPortal()(*DBusPortal,error){
 conn,err:=dbus.ConnectSessionBus()
 if err!=nil{return nil,err}
 return &DBusPortal{conn:conn},nil
}
func(p *DBusPortal)Close()error{
 if p==nil||p.conn==nil{return nil}
 return p.conn.Close()
}
func token()(string,error){
 b:=make([]byte,16)
 if _,err:=rand.Read(b);err!=nil{return "",err}
 return "ploos_"+hex.EncodeToString(b),nil
}
func(p *DBusPortal)request(ctx context.Context,method string,args ...any)(map[string]dbus.Variant,error){
 if p==nil||p.conn==nil{return nil,ErrUnavailable}
 // Install a subscription BEFORE invoking the method, since a fast
 // desktop portal may respond before the method reply arrives.
 signals:=make(chan *dbus.Signal,8)
 p.conn.Signal(signals)
 defer p.conn.RemoveSignal(signals)
 if err:=p.conn.AddMatchSignal(
  dbus.WithMatchInterface("org.freedesktop.portal.Request"),
  dbus.WithMatchMember("Response"),
  dbus.WithMatchSender(portalBus),
 );err!=nil{return nil,err}
 defer p.conn.RemoveMatchSignal(
  dbus.WithMatchInterface("org.freedesktop.portal.Request"),
  dbus.WithMatchMember("Response"),
  dbus.WithMatchSender(portalBus),
 )
 var path dbus.ObjectPath
 if err:=p.conn.Object(portalBus,portalPath).CallWithContext(ctx,method,0,args...).Store(&path);err!=nil{return nil,err}
 if !path.IsValid(){return nil,ErrPortalResponse}
 // The portal owns the request path; never accept a response from
 // another request, even if the signal has the correct interface.
 for {
  select {
  case <-ctx.Done():
   closeCtx,cancel:=context.WithTimeout(context.Background(),2*time.Second)
   _=p.conn.Object(portalBus,path).CallWithContext(closeCtx,portalRequest+".Close",0).Err
   cancel()
   return nil,ctx.Err()
  case signal,ok:=<-signals:
   if !ok{return nil,ErrPortalResponse}
   if signal==nil||signal.Path!=path||signal.Name!="org.freedesktop.portal.Request.Response"||len(signal.Body)!=2{continue}
   status,ok:=signal.Body[0].(uint32)
   if !ok{return nil,ErrPortalResponse}
   result,ok:=signal.Body[1].(map[string]dbus.Variant)
   if !ok||status!=0{return nil,ErrPortalResponse}
   return result,nil
  }
 }
}

type dbusPortalSession struct{
 portal *DBusPortal
 path dbus.ObjectPath
 mu sync.Mutex
 closed bool
 started bool
}
func(p *DBusPortal)CreateSession(ctx context.Context)(PortalSession,error){
 requestToken,err:=token();if err!=nil{return nil,err}
 sessionToken,err:=token();if err!=nil{return nil,err}
 result,err:=p.request(ctx,portalScreenCast+".CreateSession",map[string]dbus.Variant{
  "handle_token":dbus.MakeVariant(requestToken),
  "session_handle_token":dbus.MakeVariant(sessionToken),
 })
 if err!=nil{return nil,err}
 variant,ok:=result["session_handle"]
 if !ok{return nil,ErrPortalResponse}
 path,ok:=variant.Value().(dbus.ObjectPath)
 if !ok||!path.IsValid(){return nil,ErrPortalResponse}
 return &dbusPortalSession{portal:p,path:path},nil
}
func(s *dbusPortalSession)alive()error{
 s.mu.Lock();defer s.mu.Unlock()
 if s.closed{return ErrStopped}
 return nil
}
func(s *dbusPortalSession)SelectSources(ctx context.Context)error{
 if err:=s.alive();err!=nil{return err}
 t,err:=token();if err!=nil{return err}
 _,err=s.portal.request(ctx,portalScreenCast+".SelectSources",s.path,map[string]dbus.Variant{
  "handle_token":dbus.MakeVariant(t),
  "types":dbus.MakeVariant(uint32(1)), // monitor only
  "multiple":dbus.MakeVariant(false),
  "cursor_mode":dbus.MakeVariant(uint32(1)),
 })
 return err
}
func(s *dbusPortalSession)Start(ctx context.Context)error{
 if err:=s.alive();err!=nil{return err}
 t,err:=token();if err!=nil{return err}
 result,err:=s.portal.request(ctx,portalScreenCast+".Start",s.path,"",map[string]dbus.Variant{
  "handle_token":dbus.MakeVariant(t),
 })
 if err!=nil{return err}
 streams,ok:=result["streams"]
 if !ok||streams.Value()==nil{return fmt.Errorf("%w: no selected streams",ErrPortalResponse)}
 s.mu.Lock()
 defer s.mu.Unlock()
 if s.closed{return ErrStopped}
 s.started=true
 return nil
}
func(s *dbusPortalSession)Close()error{
 s.mu.Lock()
 if s.closed{s.mu.Unlock();return nil}
 s.closed=true
 s.mu.Unlock()
 ctx,cancel:=context.WithTimeout(context.Background(),2*time.Second)
 defer cancel()
 return s.portal.conn.Object(portalBus,s.path).CallWithContext(ctx,portalSession+".Close",0).Err
}

// OpenPipeWireRemote requests the portal-scoped PipeWire socket only after
// the interactive ScreenCast Start grant. The returned descriptor is local
// and MUST be closed by the caller. This method does not consume frames.
func(s *dbusPortalSession)OpenPipeWireRemote(ctx context.Context)(*os.File,error){
 s.mu.Lock()
 closed,started:=s.closed,s.started
 s.mu.Unlock()
 if closed{return nil,ErrStopped}
 if !started{return nil,ErrNotAuthorized}
 if err:=ctx.Err();err!=nil{return nil,err}
 var fd dbus.UnixFD
 err:=s.portal.conn.Object(portalBus,portalPath).
  CallWithContext(ctx,portalScreenCast+".OpenPipeWireRemote",0,
   s.path,map[string]dbus.Variant{}).Store(&fd)
 if err!=nil{return nil,err}
 if int(fd)<0{return nil,ErrPortalResponse}
 file:=os.NewFile(uintptr(fd),"ploosassist-portal-pipewire")
 if file==nil{return nil,ErrPortalResponse}
 s.mu.Lock()
 closed=s.closed
 s.mu.Unlock()
 if closed{_ =file.Close();return nil,ErrStopped}
 if err:=ctx.Err();err!=nil{_ =file.Close();return nil,err}
 return file,nil
}
