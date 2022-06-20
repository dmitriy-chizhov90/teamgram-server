package types

import (
	"testing"

	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/ntproto/convert"
)

func TestDecodeEndpoint(t *testing.T) {
	data := []struct {
		bts [16]byte
		e NtproEndpoint
		s string
	}{
		{
			[16]byte {
				2, 0, 8, 251, 42, 115, 15, 200,
				61, 247, 148, 135, 143, 12, 0, 64,
			},
			NtproEndpoint { convert.BitField128 { [2]uint64{
				14415867560673673218,
				4611699829021931325} } },
			"NewUiServerNGW1, proj: 0, api: 2, client" },
		{
			[16]byte {
				1, 240, 8, 152, 117, 86, 93, 108,
				217, 151, 83, 2, 0, 0, 0, 192,
			},
			NtproEndpoint { convert.BitField128 { [2]uint64{
				7808492391992193025,
				13835058055321196505} } },
			"AdmRlMskCln1, proj: 60, api: 1, server" },
	}

	for _, i := range data {
		e := NtproEndpoint{}
		if err := e.Decode(i.bts[0:16]); err  != nil {
			t.Errorf("Decode %v error: %v", i.bts, err)
		}
		cmpEndpoints(t, e, i.e)
		s := e.String()
		cmpStr(t, s, i.s, e)
	}
}

func TestEncodeEndpoint(t *testing.T) {
	data := []struct {
		component string
		proj, api int
		isServer bool
		e NtproEndpoint
		bts [16]byte
		s string
	}{
		{
			"NewUiServerNGW1", 0, 2, false,
			NtproEndpoint { convert.BitField128 { [2]uint64{
				14415867560673673218,
				4611699829021931325} } },
			[16]byte {
				2, 0, 8, 251, 42, 115, 15, 200,
					61, 247, 148, 135, 143, 12, 0, 64, },
			"NewUiServerNGW1, proj: 0, api: 2, client",
		},
		{
			"AdmRlMskCln1", 60, 1, true,
			NtproEndpoint { convert.BitField128 { [2]uint64{
				7808492391992193025,
				13835058055321196505} } },
			[16]byte {
				1, 240, 8, 152, 117, 86, 93, 108,
				217, 151, 83, 2, 0, 0, 0, 192, },
			"AdmRlMskCln1, proj: 60, api: 1, server",
		},
	}

	for _, i := range data {
		e := NtproEndpoint{}
		if err := e.SetComponentId(i.component); err != nil {
			t.Errorf("Wrong component id %q error: %v", i.component, err)
		}
		e.SetProjectId(i.proj)
		e.SetApiId(i.api)
		e.SetServerFlag(i.isServer)

		cmpEndpoints(t, e, i.e)

		s := e.String()
		cmpStr(t, s, i.s, e)

		var bts [16]byte
		if err := e.Encode(bts[0:16]); err  != nil {
			t.Errorf("Encode %v error: %v", bts, err)
		}

		if bts != i.bts {
			t.Errorf("%q encoded to %v, %v expected", s, bts, i.bts)
		}
	}
}

func cmpEndpoints(t *testing.T, e, ie NtproEndpoint) {
	if e.Value[0] != ie.Value[0] || e.Value[1] != ie.Value[1] {
		t.Errorf(
			"Different endpoints: %b(%d) %b(%d), expected %b(%d) %b(%d)",
			e.Value[0], e.Value[0], e.Value[1], e.Value[1],
			ie.Value[0], ie.Value[0], ie.Value[1], ie.Value[1]);
	}
}

func cmpStr(t *testing.T, s, is string, e NtproEndpoint) {
	if s != is {
		t.Errorf(
			"ToString(%b %b) = %q, expected: %q",
			e.Value[0], e.Value[1], s, is)
	}
}


