package types

import (
	"fmt"
	"encoding/binary"
)

type NtproNetgateInfo struct {
	Event NtproNetevent
	Endpoints []NtproEndpoint
}

var NetgateInfoEventId NtproEventtypeId

func init() {
	NetgateInfoEventId.Init(0, 2, 1, true)
}

func (e *NtproNetgateInfo) Init(netId* NtproEndpoint) {
	e.Event.Init(
		netId,
		netId,
		NetgateInfoEventId)
}

func (e *NtproNetgateInfo) Names() []string {
	names := make([]string, len(e.Endpoints))
	for i, point := range e.Endpoints {
		names[i] = point.String()
	}
	return names
}

func (e *NtproNetgateInfo) Decode(content []byte, event NtproNetevent) (error) {
	e.Event = event

	if e.Event.EventType != NetgateInfoEventId {
		return fmt.Errorf("received unexpected event: %s", e.Event.String())
	}

	if len(content) < 9 {
		return fmt.Errorf(
			"cannot read netgateinfo: contents len %d", len(content))
	}

	if content[0] != 1 {
		return fmt.Errorf("cannot read netgateinfo: first byte must be 1")
	}

	cnt := int(binary.LittleEndian.Uint64(content[1:9]))
    e.Endpoints = make([]NtproEndpoint, cnt)

    content = content[9:]
	if len(content) < cnt * 16 {
		return fmt.Errorf(
			"cannot read netgateinfo: %d endpoints from %d bytes",
			cnt, len(content))
	}
	
	for i := 0; i < cnt; i++ {
		if err := e.Endpoints[i].Decode(content[:16]); err != nil {
			return fmt.Errorf(
				"cannot read netgateinfo: %d-th endpoint: %v", i, err)
		}
		
		content = content[16:]
	}
	
	return nil
}

func (e *NtproNetgateInfo) Encode(b *[]byte) (error) {
	if err := e.Event.Encode(b); err != nil {
		return fmt.Errorf("cannot write netgateinfo: %v", err)
	}

	content := (*b)[NtproNeteventSize:]

	cnt := len(e.Endpoints)
	contentLen := 9 + cnt * 16
	if cap(content) < contentLen {
		return fmt.Errorf(
			"cannot write netgateinfo: contents len %d, cnt %s",
			cap(content), cnt)
    }
	
	
	content = content[:contentLen]
	content[0] = 1
	
	binary.LittleEndian.PutUint64(content[1:9], uint64(cnt))

    content = content[9:]
		
	for i, ep := range e.Endpoints {
		if err := ep.Encode(content[:16]); err != nil {
			return fmt.Errorf(
				"cannot write netgateinfo: %d-th endpoint: %v", i, err)
		}
		
		content = content[16:]
	}

	*b = (*b)[:(NtproNeteventSize + contentLen)]
	
	return nil
}



