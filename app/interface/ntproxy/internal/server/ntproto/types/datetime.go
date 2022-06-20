package types

import (
	"time"
	"fmt"
	
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/convert"
)

type NtproDatetime struct {
	convert.BitField64
}

func (e *NtproDatetime) SetNow() {
	e.Value = uint64(time.Now().UnixNano())
}

func (e *NtproDatetime) String() string {
	return fmt.Sprintf("%v", time.Unix(0, int64(e.Value)))
}


