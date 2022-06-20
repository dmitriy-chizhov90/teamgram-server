package types

import (
	"fmt"
	
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/convert"
)

type NtproNetsessionId struct {
	convert.BitField128
}

func (e *NtproNetsessionId) String() string {
	return fmt.Sprintf("%x %x", e.Value[0], e.Value[1])
}


