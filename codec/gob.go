package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// GobCodec is subclass for codec
type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

// Codec init(): var _ ReturnType = (*SubclassStruct)(parameter:nil)
// convert nil to GobCodec pointer to init() Codec class
var _ Codec = (*GobCodec)(nil)

// NewGobCodec is new
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

// ReadHeader is to decode the header
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

// ReadBody is to decode the body
func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

// Write is to encode the message
func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	// flush the buffer at the end
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err = c.enc.Encode(h); err != nil {
		log.Println("rpc: gob error encoding header:", err)
		return err
	}
	if err = c.enc.Encode(body); err != nil {
		log.Println("rpc: gob error encoding body:", err)
		return err
	}
	return
}

// Close is to close the connection
func (c *GobCodec) Close() error {
	return c.conn.Close()
}
