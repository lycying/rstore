//some codes come from : https://github.com/bsm/redeo

package codec

import (
	"bytes"
	"github.com/lycying/mut"
	"io"
	"strconv"
	"strings"
	"sync"
)

func (codec *Codec) MessageToBytes(c *mut.Conn, p mut.Packet) ([]byte, error) {
	packet := p.(*Response)
	return packet.End(), nil
}

var bufferPool sync.Pool

// Responder generates client responses
type Response struct {
	mut.Packet

	buf *bytes.Buffer
	err error
}

// NewResponder creates a new responder instance
func NewResponse() *Response {
	var buf *bytes.Buffer
	if v := bufferPool.Get(); v != nil {
		buf = v.(*bytes.Buffer)
		buf.Reset()
	} else {
		buf = new(bytes.Buffer)
	}

	return &Response{buf: buf}
}
func (r *Response) End() []byte {
	b := r.buf.Bytes()
	bufferPool.Put(r.buf)
	return b
}

// WriteBulkLen writes a bulk length
func (r *Response) WriteBulkLen(n int) {
	r.writeInline(codeBulkLen, strconv.Itoa(n))
}

// WriteBulk writes a slice
func (r *Response) WriteBulk(bulk [][]byte) {
	if r.err != nil {
		return
	}

	r.WriteBulkLen(len(bulk))
	for _, b := range bulk {
		if b == nil {
			r.WriteNil()
		} else {
			r.WriteBytes(b)
		}
	}
}

// WriteStringBulk writes a string slice
func (r *Response) WriteStringBulk(bulk []string) {
	if r.err != nil {
		return
	}

	r.WriteBulkLen(len(bulk))
	for _, b := range bulk {
		r.WriteString(b)
	}
}

// WriteString writes a bulk string
func (r *Response) WriteString(s string) {
	if r.err != nil {
		return
	}

	if err := r.buf.WriteByte(codeStrLen); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.WriteString(strconv.Itoa(len(s))); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(binCRLF); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.WriteString(s); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(binCRLF); err != nil {
		r.err = err
		return
	}
}

// WriteBytes writes a bulk string
func (r *Response) WriteBytes(b []byte) {
	if r.err != nil {
		return
	}

	if err := r.buf.WriteByte(codeStrLen); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.WriteString(strconv.Itoa(len(b))); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(binCRLF); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(b); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(binCRLF); err != nil {
		r.err = err
		return
	}
}

// WriteString writes an inline string
func (r *Response) WriteInlineString(s string) {
	r.writeInline(codeInline, s)
}

// WriteNil writes a nil value
func (r *Response) WriteNil() {
	r.writeRaw(binNIL)
}

// WriteOK writes OK
func (r *Response) WriteOK() {
	r.writeRaw(binOK)
}

// WriteInt writes an inline integer
func (r *Response) WriteInt(n int) {
	r.writeInline(codeFixnum, strconv.Itoa(n))
}

// WriteZero writes a 0 integer
func (r *Response) WriteZero() {
	r.writeRaw(binZERO)
}

// WriteOne writes a 1 integer
func (r *Response) WriteOne() {
	r.writeRaw(binONE)
}

// WriteErrorString writes an error string
func (r *Response) WriteErrorString(s string) {
	r.writeInline(codeError, s)
}

// WriteError writes an error using the standard "ERR message" format
func (r *Response) WriteError(err error) {
	s := err.Error()
	if i := strings.LastIndex(s, ": "); i > -1 {
		s = s[i+2:]
	}
	r.WriteErrorString("ERR " + s)
}

// WriteN streams data from a reader
func (r *Response) WriteN(rd io.Reader, n int64) {
	if r.err != nil {
		return
	}

	if err := r.buf.WriteByte(codeStrLen); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.WriteString(strconv.FormatInt(n, 10)); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(binCRLF); err != nil {
		r.err = err
		return
	}
	if _, err := io.CopyN(r.buf, rd, n); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(binCRLF); err != nil {
		r.err = err
		return
	}
}

// ------------------------------------------------------------------------

func (r *Response) writeInline(prefix byte, s string) {
	if r.err != nil {
		return
	}

	if err := r.buf.WriteByte(prefix); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.WriteString(s); err != nil {
		r.err = err
		return
	}
	if _, err := r.buf.Write(binCRLF); err != nil {
		r.err = err
		return
	}
}

func (r *Response) writeRaw(p []byte) {
	if r.err != nil {
		return
	}
	if _, err := r.buf.Write(p); err != nil {
		r.err = err
		return
	}
}
