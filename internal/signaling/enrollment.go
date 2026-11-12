package signaling

import (
 "errors"

 "github.com/Ploos-AS/PloosAssist/internal/session"
)

var ErrEnrollment = errors.New("role enrollment denied")

// Enrollment is a trusted in-process bridge between pairing credentials and
// verified transport identities. The transport handler MUST derive connectionID
// from its authenticated socket, never from a client-provided field.
type Enrollment struct {
 pairing *session.Store
 connections *Connections
 roles *Roles
}
func NewEnrollment(pairing *session.Store,connections *Connections,roles *Roles)(*Enrollment,error){
 if pairing==nil||connections==nil||roles==nil{return nil,ErrEnrollment}
 return &Enrollment{pairing:pairing,connections:connections,roles:roles},nil
}
func (e *Enrollment)Bind(sessionID,connectionID string,role Role,capability string)error{
 verifiedSession,_,err:=e.connections.Identity(connectionID)
 if err!=nil{return err}
 if verifiedSession!=sessionID{return ErrEnrollment}
 if role!=HostRole&&role!=TechnicianRole{return ErrEnrollment}
 expiry,err:=e.pairing.VerifyRoleCapability(sessionID,string(role),capability)
 if err!=nil{return ErrEnrollment}
 return e.roles.Bind(sessionID,connectionID,role,expiry)
}
