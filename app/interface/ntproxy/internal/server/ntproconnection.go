package server

import (
	"net"
	"fmt"
	"bufio"
	"io"
	"sync"
	"encoding/binary"
	"github.com/zeromicro/go-zero/core/logx"
)

type onConDisconnect func (e error)
type onConPack func (b []byte) error

type conRead func (r *bufio.Reader, buf *[]byte) (error)
type conWrite func (w *bufio.Writer, buf []byte) (error)

type tcpConnection struct {
	con *net.TCPConn

	reader *bufio.Reader
	writer *bufio.Writer

	writeCh chan []byte

	readBuf [128000]byte

	onDisconnect onConDisconnect
	onPack onConPack
	read conRead
	write conWrite

	onceClose sync.Once
	onceDisconnect sync.Once

	prefix string
}

var conCounter int

func (c *tcpConnection) Process(
	con *net.TCPConn,
	onDisconnect onConDisconnect,
	onPack onConPack,
	isNtpro bool,
	prefix string) {

	conCounter++
	c.prefix = fmt.Sprintf("%s %d", prefix, conCounter)
	logx.Infof("%s: created %p", c.prefix, c)

	c.con = con
	c.reader = bufio.NewReader(con)
	c.writer = bufio.NewWriter(con)

	c.writeCh = make(chan []byte)

	c.onDisconnect = onDisconnect
	c.onPack = onPack
	if isNtpro {
		c.read = readNtproPacket
		c.write = writeNtproPacket
	} else {
		c.read = plainRead
		c.write = plainWrite
	}

	go c.runRead()
	go c.runWrite()
}

func (c *tcpConnection) Write(buf []byte) (e error) {
	b := make([]byte, len(buf))
	copy(b, buf)
	
	defer func() {
		if err := recover(); err != nil {
			e = fmt.Errorf("%s: write channel closed %p", c.prefix, c)
		}
	}()
	
	c.writeCh <- b
	logx.Infof("%s: write %d bytes", c.prefix, len(b))

	return nil
}

func (c *tcpConnection) Close() {
	c.onceClose.Do(func() {
		c.con.Close()
		close(c.writeCh)
		logx.Infof("%s: connection closed", c.prefix)
	})
}

func (c *tcpConnection) runRead() {
	var err error

	defer func() {
		if err != nil {
			logx.Infof("%s: read error: %v", c.prefix, err)
			c.callOnDisconnect(err)
			c.Close()
		}
	} ()

	for {
		buf := c.readBuf[:]
		
		if err = c.read(c.reader, &buf); err != nil {
			logx.Infof("%s: read error: %v", c.prefix, err)
			return
		}
		if err = c.onPack(buf); err != nil {
			logx.Infof("%s: pack processing error: %v", c.prefix, err)
			return
		}
	}
}

func (c *tcpConnection) runWrite() {
	for b := range c.writeCh {
		if err := c.write(c.writer, b); err != nil {
			logx.Infof("%s: write error: %v", c.prefix, err)
			c.callOnDisconnect(err)
			c.Close()
		}
	}
}

func (c *tcpConnection) callOnDisconnect(err error) {
	c.onceDisconnect.Do(func() {
		logx.Infof("%s: on disconnect: %v", c.prefix, err)
		c.onDisconnect(err)
	})
}

func plainRead(r *bufio.Reader, buf *[]byte) error {
	n, err := r.Read(*buf)
	if err != nil {
		logx.Infof("plain read error: %v", err)
		return err
	}

	*buf = (*buf)[:n]
	return nil
}

func plainWrite(w *bufio.Writer, buf []byte) error {
	if _, err := w.Write(buf); err != nil {
		logx.Infof("plain write error: %v", err)
		return err
	}

	if err := w.Flush(); err != nil {
		logx.Infof("plain flush error: %v", err)
		return err
	}

	return nil
}

func readNtproPacket(r *bufio.Reader, buf *[]byte) error {
	sizeBuffer := make([]byte, 4)
	if _, err := io.ReadFull(r, sizeBuffer); err != nil {
		logx.Infof("read Ntpro pack error: %v", err)
		return err
	}

	size := int(binary.LittleEndian.Uint32(sizeBuffer))
	if size > cap(*buf) {
		logx.Infof("read Ntpro pack buffer overflow %d", size)
		return fmt.Errorf("Read buffer overflow: %d, cap: %d", size, cap(*buf))
	}

	*buf = (*buf)[:size]
	if _, err := io.ReadFull(r, *buf); err != nil {
		logx.Infof("read Ntpro pack full error: %v", err)
		return err
	}

	return nil
}

func writeNtproPacket(w *bufio.Writer, buf []byte) error {
	sizeBuffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuffer, uint32(len(buf)))
	
	if _, err := w.Write(sizeBuffer); err != nil {
		logx.Infof("write Ntpro pack size error: %v", err)
		return err
	}

	if _, err := w.Write(buf); err != nil {
		logx.Infof("write Ntpro pack error: %v", err)
		return err
	}

	if err := w.Flush(); err != nil {
		logx.Infof("write Ntpro pack flush error: %v", err)
		return err
	}

	return nil
}
