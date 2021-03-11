package codec

import (
	"io"
)

// Header includes method name, sequence number and ther error
type Header struct {
	ServiceMethod string // Service.Method
	Seq           uint64
	Error         string
}

// Codec is the interface for encode and decode message
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

// NewCodecFunc is factory for codec
type NewCodecFunc func(io.ReadWriteCloser) Codec

// Type is general name for different message format such as Gob and Json
type Type string

const (
	// GobType ...
	GobType Type = "application/gob"
	// JSONType ...
	JSONType Type = "application/json"
)

// NewCodecFuncMap service methods maps
var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
