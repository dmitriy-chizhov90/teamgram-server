package types

import (
	"fmt"
	
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/convert"
)

type NtproEventtypeId struct {
	convert.BitField64
}


func (e *NtproEventtypeId) EventId() int {
	return e.Int(0, 28)
}

func (e *NtproEventtypeId) IsFast() bool {
	return e.Int(28, 29) == 1
}

func (e *NtproEventtypeId) ApiId() int {
	return e.Int(29, 39)
}

func (e *NtproEventtypeId) ProjectId() int {
	return e.Int(39, 47)
}

func (e *NtproEventtypeId) SetEventId(id int) {
	e.SetInt(id, 0, 28)
}

func (e *NtproEventtypeId) SetFast(f bool) {
	fastFlag := 0
	if f {
		fastFlag = 1
	}
	e.SetInt(fastFlag, 28, 29)
}

func (e *NtproEventtypeId) SetApiId(id int) {
	e.SetInt(id, 29, 39)
}

func (e *NtproEventtypeId) SetProjectId(id int) {
	e.SetInt(id, 39, 47)
}

func (e *NtproEventtypeId) Init(proj, api, id int, f bool) {
	e.SetProjectId(proj)
	e.SetApiId(api)
	e.SetEventId(id)
	e.SetFast(f)
}

func (e *NtproEventtypeId) String() string {
	var fast string
	if e.IsFast() {
		fast = "fast"
	}
	return fmt.Sprintf("proj: %d, api: %d, event: %d, %s",
		e.ProjectId(), e.ApiId(), e.EventId(), fast)
}

