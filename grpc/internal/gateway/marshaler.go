package gateway

import (
	"encoding/json"
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// customJSONMarshaler implements runtime.Marshaler
type customJSONMarshaler struct{}

// Marshal marshals "v" into byte sequence.
func (j *customJSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal unmarshals "data" into "v".
func (j *customJSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads byte sequence from "r".
func (j *customJSONMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes bytes sequence into "w".
func (j *customJSONMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// ContentType returns the Content-Type which this marshaler is responsible for.
func (j *customJSONMarshaler) ContentType() string {
	return "application/json"
}
