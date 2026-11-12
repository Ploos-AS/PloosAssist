package hostclient

import (
 "bytes"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
)
func TestConsentBridgeRejectsCrossOriginAndConsumesDecision(t *testing.T){
 calls:=0
 bridge,err:=NewConsentBridge("http://127.0.0.1:9876","session",func(ticket string,approve bool)error{
  if ticket!="ticket123"||!approve{t.Fatalf("unexpected decision: %q %v",ticket,approve)}
  calls++
  return nil
 })
 if err!=nil{t.Fatal(err)}
 bridge.SetPending(PendingConsent{TicketID:"ticket123",TechnicianFingerprint:"fingerprint",ExpiresAt:time.Now().Add(time.Minute).UTC().Format(time.RFC3339)})
 send:=func(method,path,origin,token string,body []byte)int{
  t.Helper()
  req:=httptest.NewRequest(method,"http://127.0.0.1:9876"+path,bytes.NewReader(body))
  req.RemoteAddr="127.0.0.1:41234"
  if origin!=""{req.Header.Set("Origin",origin)}
  req.Header.Set("X-PloosAssist-Token",token)
  recorder:=httptest.NewRecorder()
  bridge.Handler().ServeHTTP(recorder,req)
  return recorder.Code
 }
 body,_:=json.Marshal(map[string]any{"ticketId":"ticket123","approve":true})
 if status:=send(http.MethodPost,"/api/consent/decision","http://evil.example",bridge.Token(),body);status!=403{t.Fatalf("cross origin: %d",status)}
 if status:=send(http.MethodPost,"/api/consent/decision","http://127.0.0.1:9876","wrong",body);status!=403{t.Fatalf("wrong token: %d",status)}
 if status:=send(http.MethodGet,"/api/consent","",bridge.Token(),nil);status!=200{t.Fatalf("same-origin GET: %d",status)}
 if status:=send(http.MethodPost,"/api/consent/decision","http://127.0.0.1:9876",bridge.Token(),body);status!=200{t.Fatalf("decision: %d",status)}
 if status:=send(http.MethodPost,"/api/consent/decision","http://127.0.0.1:9876",bridge.Token(),body);status!=409{t.Fatalf("replay: %d",status)}
 if calls!=1{t.Fatalf("decision called %d times",calls)}
}
