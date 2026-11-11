package signaling

import (
 "errors"
 "sync"
 "time"
)

var (
 ErrConnectionLimit = errors.New("connection capacity reached")
 ErrConnectionMissing = errors.New("connection not registered")
 ErrConnectionExpired = errors.New("connection expired")
 ErrConnectionUnauthenticated = errors.New("connection not authenticated")
)

type connectionRecord struct {
 session string
 fingerprint string
 expires time.Time
}

// Connections tracks verified, expiring transport identities.
// Registration MUST happen only after cryptographic authentication on the
// actual live transport. No client-supplied identity may be trusted directly.
type Connections struct {
 mu sync.Mutex
 records map[string]connectionRecord
 max int
 ttl time.Duration
 now func()time.Time
}
func NewConnections(max int,ttl time.Duration)(*Connections,error){
 if max<=0||ttl<=0{return nil,errors.New("invalid connection limits")}
 return &Connections{records:make(map[string]connectionRecord),max:max,ttl:ttl,now:time.Now},nil
}
func (c *Connections)Register(connectionID,sessionID,fingerprint string)error{
 if connectionID==""||sessionID==""||fingerprint==""{return ErrConnectionUnauthenticated}
 c.mu.Lock();defer c.mu.Unlock()
 now:=c.now()
 for id,rec:=range c.records{if !now.Before(rec.expires){delete(c.records,id)}}
 if _,exists:=c.records[connectionID];exists{return ErrConnectionLimit}
 if len(c.records)>=c.max{return ErrConnectionLimit}
 c.records[connectionID]=connectionRecord{session:sessionID,fingerprint:fingerprint,expires:now.Add(c.ttl)}
 return nil
}
func (c *Connections)Identity(connectionID string)(string,string,error){
 c.mu.Lock();defer c.mu.Unlock()
 rec,ok:=c.records[connectionID]
 if !ok{return "","",ErrConnectionMissing}
 if !c.now().Before(rec.expires){
  delete(c.records,connectionID)
  return "","",ErrConnectionExpired
 }
 return rec.session,rec.fingerprint,nil
}
func (c *Connections)Refresh(connectionID string)error{
 c.mu.Lock();defer c.mu.Unlock()
 rec,ok:=c.records[connectionID]
 if !ok{return ErrConnectionMissing}
 now:=c.now()
 if !now.Before(rec.expires){delete(c.records,connectionID);return ErrConnectionExpired}
 rec.expires=now.Add(c.ttl)
 c.records[connectionID]=rec
 return nil
}
func (c *Connections)Close(connectionID string){
 c.mu.Lock();defer c.mu.Unlock()
 delete(c.records,connectionID)
}
func (c *Connections)Count()int{
 c.mu.Lock();defer c.mu.Unlock()
 now:=c.now()
 for id,rec:=range c.records{if !now.Before(rec.expires){delete(c.records,id)}}
 return len(c.records)
}
