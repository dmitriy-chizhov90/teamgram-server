package convert

import (
	"fmt"
	"encoding/binary"
	"math/bits"

	//"github.com/zeromicro/go-zero/core/logx"
)

type BitField64 struct {
	Value uint64
}

type BitField128 struct {
	Value [2]uint64
}

type StdString string

func (f *BitField64) Int(low, high int) int {
	if low > high {
		return 0
	}
	var mask uint64
	mask = ^mask
	mask <<= high;

	r := f.Value &^ mask
	r >>= low
	
	return int(r)

}

func (f *BitField64) SetInt(vInt, low, high int) {
	if low > high {
		return
	}
	var mask, v uint64
	mask = ^mask
	mask >>= (64 - high + low)
	mask <<= low

	v = uint64(vInt)
	v <<= low
	
	r := f.Value &^ mask
	r |= v
	f.Value = r
}

func (f *BitField128) Int(low, high int) int {
	if high <= 64 {
		v := BitField64 { f.Value[0] }
		return v.Int(low, high)
	}
	if low >= 64 {
		v := BitField64 { f.Value[1] }
		return v.Int(low - 64, high - 64)
	}

	rawV := f.Value[0] >> low
	rawV |= (f.Value[1] << (64 - low))
	v := BitField64 { rawV }
	
	return v.Int(0, high - low)
}

func (f *BitField128) SetInt(vInt, low, high int) {
	if high <= 64 {
		bf := BitField64 { f.Value[0] }
		bf.SetInt(vInt, low, high)
		f.Value[0] = bf.Value
		return
	}
	if low >= 64 {
		bf := BitField64 { f.Value[1] }
		bf.SetInt(vInt, low - 64, high - 64)
		f.Value[1] = bf.Value
		return
	}

	lowSize := 64 - low
	highSize := high - 64
	v := uint64(vInt)
	var mask uint64
	mask = ^mask
	mask <<= lowSize
	
	lowV := v &^ mask
	highV := v >> lowSize
	bfLow, bfHigh := BitField64 { f.Value[0] }, BitField64 { f.Value[1] }
	bfLow.SetInt(int(lowV), low, 64)
	bfHigh.SetInt(int(highV), 0, highSize)

	f.Value[0], f.Value[1] = bfLow.Value, bfHigh.Value
}

func (e *BitField64) Decode (b []byte) (error) {
	if (len(b) != 8) {
		return fmt.Errorf(
			"cannot decode bitfield: incoming byte array len: %d", len(b))
	}

	e.Value = binary.LittleEndian.Uint64(b[0:8])
	return nil
}

func (e *BitField64) Encode(b []byte) (error) {
	if (len(b) != 8) {
		return fmt.Errorf(
			"cannot encode bitfield: outgoing byte array len: %d", len(b))
	}

	binary.LittleEndian.PutUint64(b[0:8], e.Value)
	return nil
}

// reverts bits in each byte in slice.
// modifies and return incoming slice
func reverseBits (bs []byte) []byte {
	for i, b := range bs {
		bs[i] = byte(bits.Reverse8(uint8(b)))
	}
	return bs
}

func (e *BitField128) Decode (b []byte) (error) {
	if (len(b) != 16) {
		return fmt.Errorf(
			"cannot decode bitfield: incoming byte array len: %d", len(b))
	}

	b = reverseBits(b)
	e.Value[0] = bits.Reverse64(binary.BigEndian.Uint64(b[0:8]))
	e.Value[1] = bits.Reverse64(binary.BigEndian.Uint64(b[8:16]))

	return nil
}

func (e *BitField128) Encode(b []byte) (error) {
	if (len(b) != 16) {
		return fmt.Errorf(
			"cannot encode bitfield: outgoing byte array len: %d", len(b))
	}

	binary.BigEndian.PutUint64(b[0:8], bits.Reverse64(e.Value[0]))
	binary.BigEndian.PutUint64(b[8:16], bits.Reverse64(e.Value[1]))
	b = reverseBits(b)
	
	return nil
}

func EncodeSixBits(c byte) (byte, error) {
    if c == ' ' {
        return 0, nil
    }
    if c >= '0' && c <= '9' {
        return (c - '0') + 1, nil
    }
    if c >= 'a' && c <= 'z' {
        return (c - 'a') + 1 + 10, nil
    }
    if c >= 'A' && c <= 'Z' {
        return (c - 'A') + 1 + 10 + 26, nil
    }

	return 0, fmt.Errorf("cannot encode six bits: %d", c)
}

func DecodeSixBits(c byte) (byte, error) {
    if c == 0 {
        return ' ', nil
    }
    if c >= 1 && c <= 10 {
        return '0' + (c - 1), nil
    }
    if c >= 11 && c <= 36 {
        return 'a' + (c - 1 - 10), nil
    }
    if c >= 37 && c <= 63 {
        return 'A' + (c - 1 - 10 - 26), nil
    }

	return 0, fmt.Errorf("cannot decode six bits: %d", c)
}

func DecodeByteSlice(b []byte) (int, []byte, error) {
	if len(b) < 8 {
		return 0, nil, fmt.Errorf(
			"cannot read size of byte slice: contents len %d", len(b))
	}

	cnt := int(binary.LittleEndian.Uint64(b[:8]))
    content := b[8:]
	if len(content) < cnt {
		return 0, nil, fmt.Errorf(
			"cannot read byte slice: %d chars from %d bytes",
			cnt, len(content))
	}

	content = content[:cnt]
	return 8 + cnt, content, nil
}

func (e *StdString) Decode (b []byte) (int, error) {
	nRead, result, err := DecodeByteSlice(b)
	if err != nil {
		return 0, fmt.Errorf("cannot read std::string: %v", err)
	}

	*e = StdString(string(result))
	return nRead, nil
}

func EncodeByteSlice(b *[]byte, data []byte) (error) {
	cnt := len(data)
	contentLen := 8 + cnt
	
	if cap(*b) < contentLen {
		return fmt.Errorf(
			"cannot write byte slice: contents len %d, cnt %d",
			cap(*b), cnt)
    }
	
	
	*b = (*b)[:contentLen]
	binary.LittleEndian.PutUint64((*b)[0:8], uint64(cnt))

    content := (*b)[8:]
		
	copy(content, data)
	return nil
}

func (e *StdString) Encode(b *[]byte) (error) {
	s := []byte(string(*e))
	return EncodeByteSlice(b, s)
}
