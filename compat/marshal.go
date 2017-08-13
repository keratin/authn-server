package compat

import "fmt"

var marshalVersion = "\x04\b" // v4.8
var marshalStringMarker = "I\""
var marshalStringTrailer = "\x06:\x06ET"

// rubyMarshal imitates Marshal.dump(str) as of v4.8
func MarshalString(str string) []byte {
	header := marshalVersion + marshalStringMarker + string(uint8(len(marshalStringTrailer)+len(str)))

	return []byte(header + str + marshalStringTrailer)
}

// rubyUnmarshal imitated Marshal.load(dumped_str) as of v4.8
func UnmarshalString(data []byte) ([]byte, error) {
	if string(data[0:2]) != marshalVersion {
		return nil, fmt.Errorf("unsupported Marshal version")
	}
	startPos := len(marshalVersion) + len(marshalStringMarker) + 1
	endPos := len(data) - len(marshalStringTrailer)
	return data[startPos:endPos], nil
}
