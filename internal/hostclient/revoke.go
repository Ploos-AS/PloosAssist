package hostclient

import "errors"

// Revoke asks the server to revoke the session using the independent host
// capability. A successful write does not prove server-side revocation:
// the server intentionally closes both WebSocket connections without an ACK.
func (c *ConsentClient)Revoke(hostCapability string)error{
 if c==nil||c.conn==nil||hostCapability==""{return errors.New("host capability required")}
 return c.conn.WriteJSON(map[string]string{"type":"session_revoke","capability":hostCapability})
}
