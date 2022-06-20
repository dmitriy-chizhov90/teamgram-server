package convert

import (
	"testing"
	"reflect"
)

func TestGetInt(t *testing.T) {
	data := []struct {
		v BitField64
		low, high int
		result int
	}{
		{ BitField64 { 0 }, 0, 64, 0 },
		{ BitField64 { 1 }, 0, 64, 1 },
		{ BitField64 { 3 }, 0, 1, 1 },
		{ BitField64 { 3 }, 1, 2, 1 },
		{ BitField64 { 2 }, 0, 1, 0 },
		{ BitField64 { 2 }, 1, 2, 1 },
		{ BitField64 { 1342177281 }, 0, 28, 1 },
		{ BitField64 { 1342177281 }, 28, 29, 1 },
		{ BitField64 { 1342177281 }, 29, 47, 2 },
	}

	for _, i := range data {
		r := i.v.Int(i.low, i.high)
		if r != i.result {
			t.Errorf(
				"%d.GetInt(%d, %d) = %d, expected: %d",
				i.v.Value, i.low, i.high, r, i.result)
		}
	}
}

func TestSetInt(t *testing.T) {
	data := []struct {
		in BitField64
		low, high int
		v int
		result BitField64
	}{
		{ BitField64 { 0 }, 0, 64, 0, BitField64 { 0 } },
		{ BitField64 { 1 }, 0, 64, 1, BitField64 { 1 } },
		{ BitField64 { 3 }, 0, 1, 1, BitField64 { 3 } },
		{ BitField64 { 3 }, 1, 2, 1, BitField64 { 3 } },
		{ BitField64 { 3 }, 2, 3, 1, BitField64 { 7 } },
		{ BitField64 { 3 }, 0, 1, 0, BitField64 { 2 } },
		{ BitField64 { 2 }, 0, 1, 0, BitField64 { 2 } },
		{ BitField64 { 2 }, 0, 1, 1, BitField64 { 3 } },
		{ BitField64 { 2 }, 1, 2, 1, BitField64 { 2 } },
		{ BitField64 { 2 }, 2, 3, 1, BitField64 { 6 } },
	}

	for _, i := range data {
		r := i.in
		r.SetInt(i.v, i.low, i.high)
		if r.Value != i.result.Value {
			t.Errorf(
				"%d.SetInt(%d, %d, %d) = %d, expected: %d",
				i.in.Value, i.v, i.low, i.high, r.Value, i.result.Value)
		}
	}
}

func TestSetInt128(t *testing.T) {
	data := []struct {
		in BitField128
		low, high int
		v int
		result BitField128
	}{
		{ BitField128 {}, 62, 66, 0, BitField128 { [2]uint64 { 0, 0 } } },
		{ BitField128 {}, 62, 66, 15, BitField128 { [2]uint64 { 13835058055282163712, 3 } } },
	}

	for _, i := range data {
		r := i.in
		r.SetInt(i.v, i.low, i.high)
		if !reflect.DeepEqual(r, i.result) {
			t.Errorf(
				"(%d %d).SetInt(%d, %d, %d) = (%d %d), expected: (%d %d)",
				i.in.Value[0], i.in.Value[1], i.v, i.low, i.high,
				r.Value[0], r.Value[1], i.result.Value[0], i.result.Value[1])
		}
	}
}

func TestDecodeEncodeString(t *testing.T) {
	data := []struct {
		in []byte
		s string
	}{
		{
			[]byte{ 38, 0, 0, 0, 0, 0, 0, 0, 123, 52, 48, 102, 56, 102, 52, 50, 53, 45, 98, 53, 57, 53, 45, 52, 101, 99, 98, 45, 57, 54, 101, 54, 45, 102, 52, 102, 99, 57, 48, 52, 49, 97, 50, 99, 57, 125 },
			"{40f8f425-b595-4ecb-96e6-f4fc9041a2c9}",
		},
	}

	for _, i := range data {
		var r StdString
		if _, err := r.Decode(i.in); err != nil {
			t.Errorf("cannot decode std::string %v: %v, expected %q", i.in, err, i.s)
		}
		if r != StdString(i.s) {
			t.Errorf("Decode(%v) = %q, expected %q", i.in, string(r), i.s)
		}

		encoded := make([]byte, 0, 100)
		if err := r.Encode(&encoded); err != nil {
			t.Errorf("cannot encode std::string %q: %v", i.s, err)
		}

		if !reflect.DeepEqual(i.in, encoded) {
			t.Errorf("Encode(%q) = %v, expected %v", string(r), encoded, i.in)
		}
	}
}



