package codec

import "errors"

const (
	codeInline  = '+'
	codeError   = '-'
	codeFixnum  = ':'
	codeStrLen  = '$'
	codeBulkLen = '*'
)

var (
	binCRLF = []byte("\r\n")
	binOK   = []byte("+OK\r\n")
	binZERO = []byte(":0\r\n")
	binONE  = []byte(":1\r\n")
	binNIL  = []byte("$-1\r\n")
)

// Protocol errors
var ErrInvalidRequest = errors.New("rstore: invalid request")
