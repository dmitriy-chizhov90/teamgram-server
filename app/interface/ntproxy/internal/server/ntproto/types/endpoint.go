package types

import (
	"fmt"
	"strings"
	
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/convert"
)

type NtproEndpoint struct {
	convert.BitField128
}

func (e *NtproEndpoint) Init(comp string, p, a int,	isServer bool) {
	e.SetComponentId(comp)
	e.SetProjectId(p)
	e.SetApiId(a)
	e.SetServerFlag(isServer)
}

func (e *NtproEndpoint) IsServer() bool {
	return e.Int(127, 128) == 1
}

func (e *NtproEndpoint) SetServerFlag(isServer bool) {
	serverFlag := 0
	if isServer {
		serverFlag = 1
	}
	e.SetInt(serverFlag, 127, 128)
}

func (e *NtproEndpoint) ApiId() int {
	return e.Int(0, 10)
}

func (e *NtproEndpoint) SetApiId(id int) {
	e.SetInt(id, 0, 10)
}

func (e *NtproEndpoint) ProjectId() int {
	return e.Int(10, 18)
}

func (e *NtproEndpoint) SetProjectId(id int) {
	e.SetInt(id, 10, 18)
}

const (
	componentIdLength = 109
	componentIdLow = 18
	componentIdHigh = componentIdLength + componentIdLow
	componentIdSize = (componentIdLength - 1) / 6
)

func (e *NtproEndpoint) IsFixed() bool {
	return e.Int(componentIdHigh - 1, componentIdHigh) == 1
}

func (e *NtproEndpoint) SetFixed(isFixed bool) {
	fixedFlag := 0
	if isFixed {
		fixedFlag = 1
	}
	
	e.SetInt(fixedFlag, componentIdHigh - 1, componentIdHigh)
}

// разворачивает срез байтов
func reverseBytes(bts []byte) []byte {
	for i, j := 0, len(bts) - 1; i < j; i, j = i + 1, j - 1 {
        bts[i], bts[j] = bts[j], bts[i]
    }
	return bts
}

func (e *NtproEndpoint) ComponentId() (string, error) {
	if !e.IsFixed() {
		return "", fmt.Errorf("Unfixed id")
	}

	bts := make([]byte, componentIdLength / 6)
	for i, _ := range bts {
		bitIndex := componentIdLow + i * 6
		code := e.Int(bitIndex, bitIndex + 6)
		
		ch, err := convert.DecodeSixBits(byte(code))
		if err != nil {
			return "", fmt.Errorf("cannot get component id: %v", err)
		}
		bts[i] = ch
	}

	bts = reverseBytes(bts)

	return strings.Trim(string(bts), " "), nil
}

func (e *NtproEndpoint) SetComponentId(s string) error {

	if len(s) > componentIdSize {
		return fmt.Errorf("Component name is too long: %d", len(s))
	}

	bts := []byte(s)
	bts = reverseBytes(bts)

	for i, ch := range bts {
		code, err := convert.EncodeSixBits(ch)
		if err != nil {
			return fmt.Errorf("cannot set component id: %v", err)
		}
		bitIndex := componentIdLow + i * 6
		e.SetInt(int(code), bitIndex, bitIndex + 6)
	}

	for i := len(bts); i < componentIdSize; i++ {
		bitIndex := componentIdLow + i * 6
		e.SetInt(0, bitIndex, bitIndex + 6)
	}

	e.SetFixed(true)
	return nil
}

func (e *NtproEndpoint) String() string {
	side := "client"
	if e.IsServer() {
		side = "server"
	}
	compId, err := e.ComponentId()
	if err != nil {
		compId = fmt.Sprintf("%s", err)
	}
	
	return fmt.Sprintf("%s, proj: %d, api: %d, %s",
		compId, e.ProjectId(), e.ApiId(), side)
}

