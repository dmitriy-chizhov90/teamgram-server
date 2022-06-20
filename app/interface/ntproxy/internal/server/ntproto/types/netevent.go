package types

import (
	"fmt"
)

type NtproHeartbeat struct {}

type NtproNetevent struct {
	Sender NtproEndpoint
	Target NtproEndpoint
	EventType NtproEventtypeId
	SessionId NtproNetsessionId
	Time NtproDatetime
	LastEndpoint NtproEndpoint
}

func (e *NtproNetevent) Init(
	sender, target *NtproEndpoint, eventTypeId NtproEventtypeId) {
	e.Sender = *sender
	e.Target = *target
	e.EventType = eventTypeId
	e.Time.SetNow()
	
	e.LastEndpoint = *sender
}

func (e *NtproNetevent) Decode(b []byte) (error) {
	if len(b) < NtproNeteventSize {
		return fmt.Errorf(
			"cannot decode NtproNetevent, incoming buf len: %d", len(b))
	}
	
	if err := e.Sender.Decode(b[0:16]); err != nil {
		return fmt.Errorf("sender decoding error : %v", err)
	}
	if err := e.Target.Decode(b[16:32]); err != nil {
		return fmt.Errorf("target decoding error : %v", err)
	}
	if err := e.EventType.Decode(b[32:40]); err != nil {
		return fmt.Errorf("eventType decoding error : %v", err)
	}
	if err := e.SessionId.Decode(b[40:56]); err != nil {
		return fmt.Errorf("sessionId decoding error : %v", err)
	}
	if err := e.Time.Decode(b[56:64]); err != nil {
		return fmt.Errorf("time decoding error : %v", err)
	}
	if err := e.LastEndpoint.Decode(b[64:80]); err != nil {
		return fmt.Errorf("lastEndpoint decoding error : %v", err)
	}
	return nil
}

func (e *NtproNetevent) Encode(b *[]byte) (error) {
	if cap(*b) < NtproNeteventSize {
		return fmt.Errorf(
			"cannot encode NtproNetevent, outgoing buf len: %d", cap(*b))
	}

	*b = (*b)[:NtproNeteventSize]
	
	if err := e.Sender.Encode((*b)[0:16]); err != nil {
		return fmt.Errorf("sender encoding error : %v", err)
	}
	if err := e.Target.Encode((*b)[16:32]); err != nil {
		return fmt.Errorf("target encoding error : %v", err)
	}
	if err := e.EventType.Encode((*b)[32:40]); err != nil {
		return fmt.Errorf("eventType encoding error : %v", err)
	}
	if err := e.SessionId.Encode((*b)[40:56]); err != nil {
		return fmt.Errorf("sessionId encoding error : %v", err)
	}
	if err := e.Time.Encode((*b)[56:64]); err != nil {
		return fmt.Errorf("time encoding error : %v", err)
	}
	if err := e.LastEndpoint.Encode((*b)[64:80]); err != nil {
		return fmt.Errorf("lastEndpoint encoding error : %v", err)
	}
	return nil
}

const (
	NtproNeteventSize = 80
)

func (e *NtproNetevent) EventId() int {
	return e.EventType.EventId()
}

func (e *NtproNetevent) ApiId() int {
	return e.EventType.ApiId()
}

func (e *NtproNetevent) ProjectId() int {
	return e.EventType.ProjectId()
}

func (e *NtproNetevent) String() string {

	return fmt.Sprintf("{%s} to {%s}, {%s} {%s} %s, last: {%s}",
		e.Sender.String(),
		e.Target.String(),
		e.EventType.String(),
		e.SessionId.String(),
		e.Time.String(),
		e.LastEndpoint.String())
}

