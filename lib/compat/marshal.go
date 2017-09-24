package compat

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var marshalVersion = "\x04\b" // v4.8
var marshalIntMarker = "i"
var marshalStringMarker = "I\""
var marshalStringTrailer = "\x06:\x06ET" // utf8 encoding?

// Marshal is an incomplete implementation of Ruby's Marshal.dump as of v4.8.
func Marshal(data interface{}) []byte {
	buf := bytes.Buffer{}
	buf.Write([]byte(marshalVersion))
	switch v := data.(type) {
	case string:
		buf.Write(marshalString(v))
	case int:
		buf.Write(marshalInt(v))
	default:
		return nil
	}
	return buf.Bytes()
}

// https://github.com/ruby/ruby/blob/d9ae8c0bc8b3798ae9e9d611a1136fc4455ff742/marshal.c#L845
func marshalString(str string) []byte {
	size := string(encodeInt(uint(len(str))))
	return []byte(marshalStringMarker + size + str + marshalStringTrailer)
}

func marshalInt(i int) []byte {
	return []byte(marshalIntMarker + string(encodeInt(uint(i))))
}

// UnmarshalString imitates Marshal.load(dumped_str) as of v4.8
func UnmarshalString(data []byte) (string, error) {
	if string(data[0:2]) != marshalVersion {
		return "", fmt.Errorf("unsupported Marshal version")
	}
	startPos := len(marshalVersion) + len(marshalStringMarker)
	endPos := len(data) - len(marshalStringTrailer)
	inner := data[startPos:endPos]

	if inner[0] < 0x05 {
		return string(inner[1+inner[0]:]), nil
	}
	return string(inner[1:]), nil
}

// UnmarshalInt imitates Marshal.load(dumped_int) as of v4.8
func UnmarshalInt(data []byte) (int, error) {
	if string(data[0:2]) != marshalVersion {
		return 0, fmt.Errorf("unsupported Marshal version")
	}
	startPos := len(marshalVersion) + len(marshalIntMarker)
	inner := data[startPos:]

	b := make([]byte, 4)
	if inner[0] < 0x05 {
		for i := uint8(0); i < inner[0]; i++ {
			b[i] = inner[i+1]
		}
	} else {
		b[0] = inner[0] - 5
	}

	var num int32
	binary.Read(bytes.NewReader(b), binary.LittleEndian, &num)
	return int(num), nil
}

// encodeInt dumps an unsigned int as of v4.8.
//
// Values up through 122 are emitted as a single byte, but shifted by 5 so that values 1, 2, 3, and
// 4 may be used for multi-byte dumps. Values greater than 122 are emitted as multiple bytes, where
// the first byte indicates how many following bytes will represent the value.
// https://github.com/ruby/ruby/blob/d9ae8c0bc8b3798ae9e9d611a1136fc4455ff742/marshal.c#L301
func encodeInt(val uint) []byte {
	if val <= 122 {
		return []byte{uint8(val + 5)}
	}

	var size uint8
	if val < 1<<8 {
		size = 1
	} else if val < 1<<16 {
		size = 2
	} else if val < 1<<32 {
		size = 3
	} else {
		size = 4
	}

	bytes := make([]byte, size+1)
	bytes[0] = size
	for i := uint8(1); i < size+1; i++ {
		bytes[i] = uint8(val & 0xff)
		val = val >> 8
	}
	return bytes
}
