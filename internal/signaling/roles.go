package signaling

import (
 "crypto/subtle"
 "errors"
 "sync"
 "time"
)

var (
 ErrRoleBinding = errors.New("invalid session role binding")
 ErrRoleOccupied = errors.New("session role already occupied")
)
type Role string
const (
 HostRole Role = "host"
 TechnicianRole Role = "technician"
)
type roleBinding struct {
 connectionID string
 fingerprint string
}
type roleSession struct {
 host roleBinding
 technician roleBinding
 expires time.Time
}

// Roles binds two verified connections to distinct session roles.
// Enrollment is restricted to a trusted orchestrator; it MUST verify the
// independent role capability before calling Bind. No client-supplied role
// string is itself proof of host or technician authority.
type Roles struct {
 mu sync.Mutex
 sessions map[string]*roleSession
 connections *Connections
 max int
 now func()time.Time
}
func NewRoles(max int,connections *Connections)(*Roles,error){
 if max<=0||connections==nil{return nil,ErrRoleBinding}
 return &Roles{sessions:make(map[string]*roleSession),connections:connections,max:max,now:time.Now},nil
}
func (r *Roles)Bind(sessionID,connectionID string,role Role,expires time.Time)error{
 if role!=HostRole&&role!=TechnicianRole||sessionID==""||!r.now().Before(expires){return ErrRoleBinding}
 verifiedSession,fingerprint,err:=r.connections.Identity(connectionID)
 if err!=nil{return err}
 if subtle.ConstantTimeCompare([]byte(verifiedSession),[]byte(sessionID))!=1{return ErrRoleBinding}
 r.mu.Lock();defer r.mu.Unlock()
 for id,s:=range r.sessions{if !r.now().Before(s.expires){delete(r.sessions,id)}}
 s,ok:=r.sessions[sessionID]
 if !ok{
  if len(r.sessions)>=r.max{return ErrConnectionLimit}
  s=&roleSession{expires:expires}
  r.sessions[sessionID]=s
 }else if !s.expires.Equal(expires){return ErrRoleBinding}
 target,other:=&s.host,&s.technician
 if role==TechnicianRole{target,other=other,target}
 if target.connectionID!=""{return ErrRoleOccupied}
 if other.connectionID==connectionID||other.fingerprint==fingerprint{return ErrRoleBinding}
 *target=roleBinding{connectionID:connectionID,fingerprint:fingerprint}
 return nil
}
func (r *Roles)Pair(sessionID string)(hostConnectionID,technicianConnectionID string,err error){
 r.mu.Lock();defer r.mu.Unlock()
 s,ok:=r.sessions[sessionID]
 if !ok||!r.now().Before(s.expires)||s.host.connectionID==""||s.technician.connectionID==""{return "","",ErrRoleBinding}
 for _,binding:=range []roleBinding{s.host,s.technician}{
  verifiedSession,fingerprint,e:=r.connections.Identity(binding.connectionID)
  if e!=nil||verifiedSession!=sessionID||fingerprint!=binding.fingerprint{return "","",ErrRoleBinding}
 }
 return s.host.connectionID,s.technician.connectionID,nil
}
func (r *Roles)Disconnect(connectionID string){
 r.mu.Lock();defer r.mu.Unlock()
 for id,s:=range r.sessions{
  if s.host.connectionID==connectionID||s.technician.connectionID==connectionID{delete(r.sessions,id)}
 }
}
func (r *Roles)Close(sessionID string){
 r.mu.Lock();defer r.mu.Unlock()
 delete(r.sessions,sessionID)
}
