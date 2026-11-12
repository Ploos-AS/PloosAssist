package hostclient

import (
 "context"
 "errors"
 "testing"

 "github.com/Ploos-AS/PloosAssist/internal/identity"
)
func TestRejectRemoteAndInvalidEndpoint(t *testing.T){
 key,err:=identity.Generate()
 if err!=nil{t.Fatal(err)}
 for _,endpoint:=range []string{"wss://127.0.0.1:8080/v1/ws","ws://example.com:8080/v1/ws","ws://192.0.2.1:8080/v1/ws","ws://127.0.0.1:8080/other","ws://127.0.0.1:8080/v1/ws?x=1"}{
  conn,_,err:=Connect(context.Background(),endpoint,"test",key)
  if conn!=nil{_ =conn.Close();t.Fatal("unexpected connection")}
  if err==nil{t.Fatalf("accepted %s",endpoint)}
 }
}
func TestRejectEmptySession(t *testing.T){
 key,_:=identity.Generate()
 _,_,err:=Connect(context.Background(),"ws://127.0.0.1:8080/v1/ws","",key)
 if !errors.Is(err,ErrProtocol){t.Fatalf("empty session: %v",err)}
}
