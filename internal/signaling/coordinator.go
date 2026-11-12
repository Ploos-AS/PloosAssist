package signaling

import (
 "errors"
 "time"
)

var ErrHostIdentity = errors.New("host identity does not match authenticated connection")

// Coordinator binds consent operations to the verified identity of a live
// connection. It does not itself prove that a human clicked an approval UI.
type Coordinator struct {
 connections *Connections
 consent *ConsentPolicy
}
func NewCoordinator(connections *Connections,consent *ConsentPolicy)(*Coordinator,error){
 if connections==nil||consent==nil{return nil,ErrConsentInvalid}
 return &Coordinator{connections:connections,consent:consent},nil
}

// Enroll is a trusted orchestration operation: callers must have independently
// established which authenticated connection belongs to the host and which
// belongs to the technician, including the host's intent to pair.
func (c *Coordinator)Enroll(sessionID,hostConnectionID,techConnectionID string,expires time.Time)error{
 hostSession,host,err:=c.connections.Identity(hostConnectionID)
 if err!=nil{return err}
 techSession,tech,err:=c.connections.Identity(techConnectionID)
 if err!=nil{return err}
 if sessionID==""||hostSession!=sessionID||techSession!=sessionID||host==tech{return ErrConsentInvalid}
 return c.consent.Create(sessionID,host,tech,expires)
}

// ApproveHost must be called only as the result of an explicit, trusted host
// UI action on the actual authenticated host connection.
func (c *Coordinator)ApproveHost(sessionID,hostConnectionID,techConnectionID string)error{
 hostSession,host,err:=c.connections.Identity(hostConnectionID)
 if err!=nil{return err}
 techSession,tech,err:=c.connections.Identity(techConnectionID)
 if err!=nil{return err}
 if hostSession!=sessionID||techSession!=sessionID{return ErrHostIdentity}
 return c.consent.Approve(sessionID,host,tech)
}
func (c *Coordinator)RevokeHost(sessionID,hostConnectionID string)error{
 hostSession,host,err:=c.connections.Identity(hostConnectionID)
 if err!=nil{return err}
 if hostSession!=sessionID{return ErrHostIdentity}
 return c.consent.Revoke(sessionID,host)
}
