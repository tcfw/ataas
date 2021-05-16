package marshalers

import (
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

// JSONMarshaler is a Marshaler which marshals/unmarshals into/from JSON
// with the standard "encoding/json" package of Golang.
// Although it is generally faster for simple proto messages than JSONPb,
// it does not support advanced features of protobuf, e.g. map, oneof, ....
//
// The NewEncoder and NewDecoder types return *json.Encoder and
// *json.Decoder respectively.
type JSONMarshaler struct{}

// ContentType always Returns "application/json".
func (*JSONMarshaler) ContentType() string {
	return "application/json"
}

// Marshal marshals "v" into JSON
func (j *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal unmarshals JSON data into "v".
func (j *JSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads JSON stream from "r".
func (j *JSONMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *JSONMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// Delimiter for newline encoded JSON streams.
func (j *JSONMarshaler) Delimiter() []byte {
	return []byte("\n")
}
