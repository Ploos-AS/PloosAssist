package hostclient

import (
 "encoding/json"
 "errors"
 "time"

 "github.com/gorilla/websocket"
)

var ErrConsentResponse = errors.New("invalid host consent response")

// ConsentClient wraps the authenticated host WebSocket. It must be used
// sequentially by its owning host process, not shared across readers.
type ConsentClient struct {conn *websocket.Conn}
type PendingConsent struct {
 TicketID string `json:"ticket_id"`
 TechnicianFingerprint string `json:"technician_fingerprint"`
 ExpiresAt string `json:"expires_at"`
}
func NewConsentClient(conn *websocket.Conn)*ConsentClient{return &ConsentClient{conn:conn}}

// EnrollHost uses a capability from the local pairing workflow, after Connect
// has authenticated the host's Ed25519 identity.
func (c *ConsentClient)EnrollHost(capability string)error{
 if c==nil||c.conn==nil||capability==""{return ErrConsentResponse}
 if err:=c.conn.WriteJSON(map[string]string{"type":"role_enroll","role":"host","capability":capability});err!=nil{return err}
 var reply struct{Type string `json:"type"`;Role string `json:"role"`}
 if err:=c.conn.ReadJSON(&reply);err!=nil{return err}
 if reply.Type!="role_enrolled"||reply.Role!="host"{return ErrConsentResponse}
 return nil
}
func (c *ConsentClient)Check()(PendingConsent,bool,error){
 if c==nil||c.conn==nil{return PendingConsent{},false,ErrConsentResponse}
 if err:=c.conn.WriteJSON(map[string]string{"type":"approval_check"});err!=nil{return PendingConsent{},false,err}
 var raw json.RawMessage
 if err:=c.conn.ReadJSON(&raw);err!=nil{return PendingConsent{},false,err}
 var kind struct{Type string `json:"type"`}
 if err:=json.Unmarshal(raw,&kind);err!=nil{return PendingConsent{},false,err}
 if kind.Type=="approval_none"{return PendingConsent{},false,nil}
 if kind.Type!="approval_pending"{return PendingConsent{},false,ErrConsentResponse}
 var pending PendingConsent
 if err:=json.Unmarshal(raw,&pending);err!=nil{return PendingConsent{},false,err}
 expiry,err:=time.Parse(time.RFC3339,pending.ExpiresAt)
 if err!=nil||pending.TicketID==""||pending.TechnicianFingerprint==""||!time.Now().Before(expiry){return PendingConsent{},false,ErrConsentResponse}
 return pending,true,nil
}
func (c *ConsentClient)Decide(ticketID string,approve bool)error{
 if c==nil||c.conn==nil||ticketID==""{return ErrConsentResponse}
 if err:=c.conn.WriteJSON(map[string]any{"type":"approval_decide","ticket_id":ticketID,"approve":approve});err!=nil{return err}
 var reply struct{Type string `json:"type"`;State string `json:"state"`}
 if err:=c.conn.ReadJSON(&reply);err!=nil{return err}
 expected:="denied"
 if approve{expected="approved"}
 if reply.Type!="approval_decided"||reply.State!=expected{return ErrConsentResponse}
 return nil
}
