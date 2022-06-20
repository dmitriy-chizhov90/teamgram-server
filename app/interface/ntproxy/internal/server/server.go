// Copyright 2022 Teamgram Authors
//  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: teamgramio (teamgram.io@gmail.com)
//

package server

import (
	"net"
	"fmt"
	"github.com/teamgram/marmota/pkg/timer2"
	"github.com/zeromicro/go-zero/core/logx"
	
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/config"
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/types"
)

type Server struct {
	c        *config.Config
	timer    *timer2.Timer // 32 * 2048
	l        net.Listener
	isClosed bool
}

func New(c config.Config) *Server {
	var (
		s = new(Server)
	)

	s.c = &c
	s.timer = timer2.NewTimer(1024)
	
	logx.Infof("New: %v, %s", c, netEndpoint.String())

	return s
}

func (s *Server) Serve() (e error) {

	logx.Infof("Serve")

	if len(s.c.Server.Addrs) != 1 {
		return fmt.Errorf("Only one server address suported")
	}

	s.l, e = net.Listen("tcp", s.c.Server.Addrs[0])
    if e != nil {
        return e
    }

	go func() {
		for i := 0; ; i++ {
			conn, err := s.l.Accept()
			if err != nil {
				if s.isClosed {
					break;
				}
				logx.Errorf("Connection cannot be accepted: %v", err)
				continue
			}
			
			handleNtproConn(conn.(*net.TCPConn), i, s.c, processNtproBusEvent)
		}
	} ()
	
	return nil
}

func (s *Server) Close() {
	logx.Infof("Close")

	if s.l != nil {
		s.isClosed = true
		s.l.Close()
	}
}

type Encoder interface {
	Encode(buf *[]byte) (err error)
}

func processNtproBusEvent(
	c *ntproConnection,
	e types.NtproNetevent,
	buf []byte) (error) {
	s, isConnected := c.sessions[e.SessionId]
	if !isConnected {
		logx.Infof("%s: in pack for unconnected session: %v, %v",
			c.prefix, e, buf)
		return fmt.Errorf("bus event for unconnected session")
	}

	switch e.EventType {
	case types.ConnectChatEventId:
		return c.processConnectChatEvent(e, buf, s)
	case types.DisconnectChatEventId:
		return c.processDisconnectChatEvent(e, buf, s)
	case types.SendChatMsgEventId:
		return c.processSendChatMsgEvent(e, buf, s)
	default:
		logx.Infof("%s: in pack: %v %v", c.prefix, e, buf)
	}

	return nil
}

func (c *ntproConnection) processConnectChatEvent(
	e types.NtproNetevent,
	buf []byte,
	s *ntproSession) (error) {

	var subEvent types.NtproSubscribeChat
	subEvent.Event = e
	if err := subEvent.Decode(buf); err != nil {
		return fmt.Errorf("cannot process ntpro subscribe chat event %v: %v, %v",
			e, err, buf)
	}

	id := string(subEvent.Id)
	if _, alreadyExists := s.subscriptions[id]; alreadyExists {
		return fmt.Errorf("subscription already exists %s", id)
	}

	sub := &ntproSubscription {	subEvent: subEvent, session: s }
	s.subscriptions[id] = sub

	go connectToChat(c, sub)
	return nil
}

func (c *ntproConnection) processDisconnectChatEvent(
	e types.NtproNetevent,
	buf []byte,
	s *ntproSession) (error) {

	var subEvent types.NtproUnsubscribeChat
	subEvent.Event = e
	if err := subEvent.Decode(buf); err != nil {
		return fmt.Errorf("cannot process ntpro unsubscribe chat event %v: %v, %v",
			e, err, buf)
	}

	id := string(subEvent.Id)
	if sub, alreadyExists := s.subscriptions[id]; alreadyExists {
		sub.close();
		delete(s.subscriptions, id)
	}

	return nil
}

func (c *ntproConnection) processSendChatMsgEvent(
	e types.NtproNetevent,
	buf []byte,
	s *ntproSession) (error) {

	var subEvent types.NtproSendChatMsg
	subEvent.Event = e
	if err := subEvent.Decode(buf); err != nil {
		return fmt.Errorf("cannot process ntpro send chat msg event %v: %v, %v",
			e, err, buf)
	}

	id := string(subEvent.Id)
	sub, found := s.subscriptions[id]
	if !found {
		return fmt.Errorf("subscription not found for event %v", subEvent)
	}

	if err := sub.chatCon.Write(subEvent.Buf); err != nil {
		return fmt.Errorf("cannot send msg to chat: %v", err)
	}

	return nil
}

func connectToChat(
	c *ntproConnection,
	sub *ntproSubscription) {

	// Открываем соединение.
	conn, err := net.Dial("tcp", "localhost:10443")
	if err != nil {
		logx.Infof("Cannot establish chat connection for %v", *sub)
		sendChatDisconnected(c, sub)
		return 
	}

	onDsc := func(err error) { processChatDisconnected(c, sub, err) }
	onPack := func(b []byte) error { return readChatMsg(c, sub, b) }
	
	sub.chatCon.Process(conn.(*net.TCPConn), onDsc, onPack, false, "chat con")

	if err = sendChatConnected(c, sub); err != nil {
		logx.Infof("Cannot send connected event: %v", err)
		sendChatDisconnected(c, sub)
	}
}

func sendChatConnected(c *ntproConnection, sub *ntproSubscription) error {
	e := sub.subEvent.NewChatConnectedEvent()
	if err := sendChatEventToNtpro(c, &e); err != nil {
		return err
	}
	logx.Infof("%s: send chat connected: %v", c.prefix, e)
	return nil
}


func sendChatDisconnected(c *ntproConnection, sub *ntproSubscription) {
	e := sub.subEvent.NewChatDisconnectedEvent()
	sub.close()
	
	if err := sendChatEventToNtpro(c, &e); err != nil {
		logx.Infof("%s: cannot send disconnect event: %v", c.prefix, err)
	}
}

func sendChatEventToNtpro(c *ntproConnection, e Encoder) error {
	buf := c.wBuf[:0]
	if err := e.Encode(&buf); err != nil {
		return err
	}
	return c.Write(buf)
}

func processChatDisconnected(
	c *ntproConnection,
	sub *ntproSubscription,
	err error) {

	logx.Infof("%s: chat disconnected: %v", sub.chatCon.prefix, err)
	sendChatDisconnected(c, sub)
}

func readChatMsg(c *ntproConnection, sub *ntproSubscription, buf []byte) error {
	e := sub.subEvent.NewChatMsgReceivedEvent(buf)
	
	if err := sendChatEventToNtpro(c, &e); err != nil {
		logx.Infof("%s: cannot send received from chat msg: %v", c.prefix, err)
		return err
	}
	return nil
}





