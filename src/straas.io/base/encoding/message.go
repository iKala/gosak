package encoding

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	_        = iota
	typeJSON = iota
	typeYAML = iota
)

// RawMessage wrapper both json and yaml raw message type
// (https://golang.org/pkg/encoding/json/#RawMessage)
// RawMessage is inspired by json.RawMessage but this is able
// to handle both JSON and YAML
type RawMessage struct {
	// for JSON
	data []byte
	// for YAML
	unmarshaler func(interface{}) error
	// codec indicates the codec
	// TODO: use iota
	codec int32
}

// To converts raw message to the given type
func (m *RawMessage) To(v interface{}) error {
	switch m.codec {
	case typeJSON:
		return json.Unmarshal(m.data, v)
	case typeYAML:
		return m.unmarshaler(v)
	default:
		return fmt.Errorf("unknown codec %v", m.codec)
	}
}

// UnmarshalJSON sets data to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	m.codec = typeJSON
	m.data = append(m.data[0:0], data...)
	return nil
}

// UnmarshalYAML implements the interface of yaml.Unmarshaler
func (m *RawMessage) UnmarshalYAML(unmarshaler func(interface{}) error) error {
	m.codec = typeYAML
	m.unmarshaler = unmarshaler
	return nil
}
