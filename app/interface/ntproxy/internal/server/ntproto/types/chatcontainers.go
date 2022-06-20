package types

import (
	"fmt"

	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/convert"
)

type StdStringEvent struct {
	Event NtproNetevent
	Id convert.StdString }

type BufferEvent struct {
	Event NtproNetevent
	Id convert.StdString
	Buf []byte }

type NtproSubscribeChat struct {StdStringEvent}
type NtproUnsubscribeChat struct {StdStringEvent}
type NtproChatConnected struct {StdStringEvent}
type NtproChatDisconnected struct {StdStringEvent}
type NtproSendChatMsg struct {BufferEvent}
type NtproReceiveChatMsg struct {BufferEvent}

var ConnectChatEventId NtproEventtypeId
var DisconnectChatEventId NtproEventtypeId
var SendChatMsgEventId NtproEventtypeId

var ChatConnectedEventId NtproEventtypeId
var ChatDisconnectedEventId NtproEventtypeId
var ChatMsgReceivedEventId NtproEventtypeId

func init() {
	ConnectChatEventId.Init(60, 3, 1, true)
	DisconnectChatEventId.Init(60, 3, 2, true)
	SendChatMsgEventId.Init(60, 3, 3, true)

	ChatConnectedEventId.Init(60, 3, 4, true)
	ChatDisconnectedEventId.Init(60, 3, 5, true)
	ChatMsgReceivedEventId.Init(60, 3, 6, true)
}

func (e *NtproSubscribeChat) NewChatConnectedEvent() NtproChatConnected {
	r :=  NtproChatConnected { StdStringEvent {
		Event: e.Event,
		Id: e.Id }}
	r.Event.Init(&e.Event.Target, &e.Event.Sender, ChatConnectedEventId)
	return r
}

func (e *NtproSubscribeChat) NewChatDisconnectedEvent() NtproChatDisconnected {
	r :=  NtproChatDisconnected { StdStringEvent {
		Event: e.Event,
			Id: e.Id }}
	r.Event.Init(&e.Event.Target, &e.Event.Sender, ChatDisconnectedEventId)
	return r
}

func (e *NtproSubscribeChat) NewChatMsgReceivedEvent(b []byte) NtproReceiveChatMsg {
	r :=  NtproReceiveChatMsg { BufferEvent {
		Event: e.Event,
		Id: e.Id,
		Buf: b}}
	r.Event.Init(&e.Event.Target, &e.Event.Sender, ChatMsgReceivedEventId)
	return r
}

func (e *StdStringEvent) Decode(content []byte) (error) {
	ePrefix := "cannot read std::string event"
	if len(content) < 1 {
		return fmt.Errorf("%s: contents len %d", ePrefix, len(content))
	}

	if content[0] != 1 {
		return fmt.Errorf("%s: first byte must be 1", ePrefix)
	}

	if _, err := e.Id.Decode(content[1:]); err != nil {
		return fmt.Errorf("%s: first byte must be 1", ePrefix)
	}

	return nil
}

func (e *StdStringEvent) Encode(buf *[]byte) (error) {
	ePrefix := "cannot write std::string event"
	
	if err := e.Event.Encode(buf); err != nil {
		return fmt.Errorf("%s: %v", ePrefix, err)
	}

	content := (*buf)[NtproNeteventSize:]
	if cap(content) < 1 {
		return fmt.Errorf("%s: contents len %d", ePrefix, cap(content))
    }
	content = content[:1]
	content[0] = 1
	
	content = content[1:]
	if err := e.Id.Encode(&content); err != nil {
		return fmt.Errorf("%s: %v", ePrefix, err)
	}
	
	*buf = (*buf)[:(NtproNeteventSize + len(content) + 1)]
	
	return nil
}

func (e *BufferEvent) Decode(content []byte) (error) {
	ePrefix := "cannot read buffer event"
	if len(content) < 1 {
		return fmt.Errorf("%s: contents len %d", ePrefix, len(content))
	}

	if content[0] != 1 {
		return fmt.Errorf("%s: first byte must be 1", ePrefix)
	}

	content = content[1:]
	idLen, err := e.Id.Decode(content)
	if err != nil {
		return fmt.Errorf("%s: during id decoding: %v", ePrefix, err)
	}

	content = content[idLen:]
	_, buf, err := convert.DecodeByteSlice(content)
	if err != nil {
		return fmt.Errorf("%s: during buf decoding: %v", ePrefix, err)
	}

	e.Buf = buf
	return nil
}

func (e *BufferEvent) Encode(buf *[]byte) (error) {
	ePrefix := "cannot write buffer event"
	
	if err := e.Event.Encode(buf); err != nil {
		return fmt.Errorf("%s: %v", ePrefix, err)
	}

	size := NtproNeteventSize
	content := (*buf)[size:]
	if cap(content) < 1 {
		return fmt.Errorf("%s: contents len %d", ePrefix, cap(content))
    }
	content = content[:1]
	content[0] = 1

	size += 1
	content = content[1:]
	
	if err := e.Id.Encode(&content); err != nil {
		return fmt.Errorf("%s: %v", ePrefix, err)
	}

	size += len(content)
	content = content[len(content):]
	
	if err := convert.EncodeByteSlice(&content, e.Buf); err != nil {
		return fmt.Errorf("%s: %v", ePrefix, err)
	}
	
	size += len(content)
	*buf = (*buf)[:size]
	
	return nil
}




