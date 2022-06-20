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
	"encoding/binary"
	"github.com/zeromicro/go-zero/core/logx"
	
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/config"
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/types"
)

const (
	chatCompId = "ChatProxy"
	chatProj = 60
	chatApi = 3
	chatIsServer = true
	
	netCompId = "ChatNetgate"
	netProj = 0
	netApi = 2
	netIsServer = true

	heartbeatData = 13
)

var (
	netEndpoint types.NtproEndpoint
	chatEndpoint types.NtproEndpoint
)

type onNtproEvent func (c *ntproConnection,	e types.NtproNetevent, buf []byte) error

type ntproConnection struct {
	tcpConnection

	index int
	info ntproConnectionInfo

	wBuf [128000]byte

	sessions map[types.NtproNetsessionId]*ntproSession

	state int

	onEvent onNtproEvent
}

const (
	initial = iota
	handshake
	initialized
)

// Соединение Ntpro api в рамках одного сетевого подключения
type ntproSession struct {
	conEvent types.NtproNetgateConnected
	con *ntproConnection
	subscriptions map[string]*ntproSubscription
}

// Подписка на чат из клиентского приложения
type ntproSubscription struct {
	subEvent types.NtproSubscribeChat
	session *ntproSession
	chatCon tcpConnection
}

// Сетевое подключение между сервером NTPro и сервером чата
type ntproConnectionInfo struct {
	serviceId int
	channelType int
}

func init() {
	netEndpoint.Init(netCompId, netProj, netApi, netIsServer)
	chatEndpoint.Init(chatCompId, chatProj, chatApi, chatIsServer)
}

func handleNtproConn(
	c *net.TCPConn,
	conIndex int,
	cfg *config.Config,
	onEvent onNtproEvent) {

	logx.Infof("New incoming ntproConnection %d", conIndex)

	con := ntproConnection {
		index: conIndex,
		sessions: make(map[types.NtproNetsessionId]*ntproSession),
		onEvent: onEvent,
		state: initial}

	onPack := func(buf []byte) error {
		if err := con.processNtproPack(buf, cfg.ServiceId); err != nil {
			logx.Infof("con %d: cannot process pack: %v", conIndex, err)
			return err
		}
		return nil
	}

	onDsc := func(e error) { processNtproDisconnect(&con, e) }
	
	con.Process(c, onDsc, onPack, true, "ntpro con")
}

func processNtproDisconnect(c *ntproConnection,	err error) {
	for _, session := range c.sessions {
		for _, sub := range session.subscriptions {
			sub.close()
		}
	}
	c.Close()
}

func (c *ntproConnection) processNtproPack(buf []byte, serviceId int) error {
	switch c.state {
	case initial:
		if err := c.initConnection(buf, serviceId); err != nil {
			logx.Infof("Connection %d init error: %v", c.index, err)
			return err
		}
		c.state = handshake
		logx.Infof("Con %d handshake received", c.index)
	case handshake:
		if err := c.acceptHandshake(buf); err != nil {
			logx.Infof("%s handshake accepting error: %v", c.prefix, err)
			return err
		}
		c.state = initialized
		logx.Infof("Con %d initialized", c.index)
	default:
		if err := c.readNtproEvent(buf); err != nil {
			logx.Infof("%s: read ntpro event error: %v", c.prefix, err)
			break
		}
	}

	return nil
}

func (c *ntproConnection) initConnection(buf []byte, serviceId int) error {

	/// read handshake
	conInfo, err := c.readHandshake1(buf)
	if err != nil {
		return err
	}

	c.info = *conInfo
	logx.Infof("%s: receive handhsake: %v", c.prefix, conInfo)

	// send handshake
	if err = c.writeHandshake(serviceId); err != nil {
		return err
	}
	
	return nil
}

func (c *ntproConnection) acceptHandshake(buf []byte) error {
	// read handshake ack
	if err := c.readHandshake2(buf); err != nil {
		return err
	}

	if err := c.writeNetgateInfo(); err != nil {
		return err
	}

	return nil
}

func (c *ntproConnection) readHandshake1(buf []byte) (*ntproConnectionInfo, error) {
	handshakeSize := 16
	if (len(buf) != handshakeSize)	{
		return nil, fmt.Errorf("cannot read handshake1: expected %d bytes but %d got", handshakeSize, len(buf))
	}

	msgType := 1
	if mt := int(binary.LittleEndian.Uint32(buf[0:4])); mt != msgType {
		return nil, fmt.Errorf("cannot read handshake1: expected msgType %d but %d got", msgType, mt)
	}
	serviceId := int(binary.LittleEndian.Uint64(buf[4:12]))
	channelType := int(binary.LittleEndian.Uint32(buf[12:16]))
	
	ci := ntproConnectionInfo { serviceId: serviceId, channelType: channelType }
	return &ci, nil
}

func (c *ntproConnection) readHandshake2(buf []byte) (error) {
	handshakeSize := 5
	if (len(buf) != handshakeSize)	{
		return fmt.Errorf("cannot read handshake2: expected %d bytes but %d got", handshakeSize, len(buf))
	}

	msgType := 3
	if mt := int(binary.LittleEndian.Uint32(buf[0:4])); mt != msgType {
		return fmt.Errorf("cannot read handshake2: expected msgType %d but %d got", msgType, mt)
	}
	if 1 != buf[4] {
		return fmt.Errorf("handshake rejected by remote side")
	}

	return nil
}

func (c *ntproConnection) readNtproEvent(buf []byte) (error) {
	if len(buf) == 1 && buf[0] == heartbeatData {
		if err := c.writeHeartbeat(); err != nil {
			logx.Infof("Con %d cannot write handshake: %v", c.index, err)
			return err
		}
		return nil
	}
	
	e := types.NtproNetevent {}
	if err := e.Decode(buf); err != nil {
		return err
	}

	content := buf[types.NtproNeteventSize:]

	logx.Infof("con %d: in event: %s, %d bytes",
		c.index, e.String(), len(content))

	switch e.EventType {
	case types.NetgateInfoEventId:
		return c.processNetgateInfo(e, content)
	case types.ConnectedEventId:
		return c.processConEvent(e)
	case types.DisconnectedEventId:
		return c.processDisconEvent(e)
	default:
		return c.onEvent(c, e, content)
	}
	
	return nil
}

func (c *ntproConnection) writeHeartbeat() (error) {
	wBuf := c.wBuf[:1]
	wBuf[0] = heartbeatData
	logx.Infof("con %d: write heartbeat", c.index)
	return c.Write(wBuf)
}

func (c *ntproConnection) processNetgateInfo(e types.NtproNetevent, buf []byte) (error) {
	info := types.NtproNetgateInfo {}
	if err := info.Decode(buf, e); err != nil {
		return err
	}
	
	logx.Infof("con %d: in netgate info: %v, %v", c.index, e, info.Names())

	return nil
}

func (c *ntproConnection) processConEvent(eRaw types.NtproNetevent) (error) {
	e := types.NtproNetgateConnected{eRaw}
		
	logx.Infof("con %d: in session connected: %v", c.index, e)

	if e.Target == chatEndpoint {
		c.sessions[e.SessionId] = &ntproSession{
			conEvent: e,
			con: c,
			subscriptions: make(map[string]*ntproSubscription)}
		logx.Infof("con %d: new chat session: %v", c.index, e)
	}

	return nil
}

func (sub *ntproSubscription) close() {
	sub.chatCon.Close()
	delete(sub.session.subscriptions, string(sub.subEvent.Id))
}

func (c *ntproConnection) processDisconEvent(eRaw types.NtproNetevent) (error) {
	e := types.NtproNetgateConnected{eRaw}
	session, ok := c.sessions[e.SessionId]
	if !ok {
		return nil
	}

	for _, sub := range session.subscriptions {
		sub.close()
	}
	delete(c.sessions, e.SessionId)
	
	logx.Infof("con %d: in session disconnected: %v", c.index, e)

	return nil
}

func (c *ntproConnection) writeHandshake(serviceId int) error {
	handshakeSize := 17

	if cap(c.wBuf) < handshakeSize {
		return fmt.Errorf("cannot write handshake: not enough size")
	}

	buf := c.wBuf[:handshakeSize]
	msgType := uint32(2)
	channelType := uint32(0)
	isOk := byte(1)
	
	binary.LittleEndian.PutUint32(buf[0:4], msgType)
	binary.LittleEndian.PutUint64(buf[4:12], uint64(serviceId))
	binary.LittleEndian.PutUint32(buf[12:16], channelType)
	buf[16] = isOk

	return c.Write(buf)
}

func (c *ntproConnection) writeNetgateInfo() error {
	info := types.NtproNetgateInfo {}
	
	info.Init(&netEndpoint)
	info.Endpoints = append(info.Endpoints, chatEndpoint)

	buf := c.wBuf[:0]
	
	if err := info.Encode(&buf); err != nil {
		return err
	}
	
	logx.Infof("out netgate info: %s, %v, %d",
		info.Event.String(), info.Names(), len(buf))
	
	return c.Write(buf)
}


