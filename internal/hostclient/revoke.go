package hostclient

import (
 "errors"
 "time"
)

// Revoke sends the independent host capability and waits for the server's
// explicit acknowledgment that the in-memory revocation has committed.
func (c *ConsentClient)Revoke(hostCapability string)error{
 if c==nil||c.conn==nil||hostCapability==""{return errors.New("host capability required")}
 if err:=c.conn.WriteJSON(map[string]string{"type":"session_revoke","capability":hostCapability});err!=nil{return err}
 if err:=c.conn.SetReadDeadline(time.Now().Add(5*time.Second));err!=nil{return err}
 var reply struct{Type string `json:"type"`}
 if err:=c.conn.ReadJSON(&reply);err!=nil{return err}
 if reply.Type!="session_revoked"{return errors.New("unexpected revocation response")}
 return nil
}
