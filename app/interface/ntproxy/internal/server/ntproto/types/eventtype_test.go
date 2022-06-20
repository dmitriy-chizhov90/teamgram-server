package types

import (
	"testing"

	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/convert"
)

func TestDecodeEventtype(t *testing.T) {
	data := []struct {
		bts [8]byte
		e NtproEventtypeId
		s string
	}{
		{
			[8]byte {
				1, 0, 0, 80, 0, 0, 0, 0,
			},
			NtproEventtypeId { convert.BitField64 {
				1342177281 } },
			"proj: 0, api: 2, event: 1, fast" },
	}

	for _, i := range data {
		e := NtproEventtypeId{}
		if err := e.Decode(i.bts[0:8]); err  != nil {
			t.Errorf("Decode %v error: %v", i.bts, err)
		}
		cmpEventtypes(t, e, i.e)
		s := e.String()
		cmpStr(t, s, i.s, e)
	}
}

func TestEncodeEventtype(t *testing.T) {
	data := []struct {
		proj, api, event int
		f bool
		e NtproEventtypeId
		bts [8]byte
		s string
	}{
		{
			0, 2, 1, true,
			NtproEventtypeId { convert.BitField64 { 
				1342177281 } },
			[8]byte {
				1, 0, 0, 80, 0, 0, 0, 0,
			},
			"proj: 0, api: 2, event: 1, fast",
		},
	}

	for _, i := range data {
		e := NtproEventtypeId{}
		e.SetProjectId(i.proj)
		e.SetApiId(i.api)
		e.SetEventId(i.event)
		e.SetFast(i.f)

		cmpEventtypes(t, e, i.e)

		s := e.String()
		cmpStr(t, s, i.s, e)

		var bts [8]byte
		if err := e.Encode(bts[0:8]); err  != nil {
			t.Errorf("Encode %v error: %v", bts, err)
		}

		if bts != i.bts {
			t.Errorf("%q encoded to %v, %v expected", s, bts, i.bts)
		}
	}
}

func cmpEventtypes(t *testing.T, e, ie NtproEventtypeId) {
	if e.Value != ie.Value {
		t.Errorf(
			"Different endpoints: %b(%d), expected %b(%d)",
			e.Value, e.Value, ie.Value, ie.Value);
	}
}

func cmpStr(t *testing.T, s, is string, e NtproEventtypeId) {
	if s != is {
		t.Errorf(
			"ToString(%bdmitry.chizhov) = %q, expected: %q",
			e.Value, s, is)
	}
}


