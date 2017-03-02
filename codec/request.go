package codec

import (
	"github.com/lycying/mut"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	mut.Packet
	C string
	P []string
}

func (req *Request) ParamsLen() int {
	return len(req.P)
}

func (codec *Codec) ReadPacket(c *mut.Conn) (mut.Packet, error) {
	line, err := c.ReadString('\n')
	if err != nil || len(line) < 3 {
		return nil, io.EOF
	}

	// Truncate CRLF
	line = line[:len(line)-2]

	// Return if inline
	if line[0] != codeBulkLen {
		return &Request{C: strings.ToLower(line)}, nil
	}

	argc, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, ErrInvalidRequest
	}

	args := make([]string, argc)
	for i := 0; i < argc; i++ {
		if args[i], err = codec.parseArgument(c); err != nil {
			return nil, err
		}
	}
	return &Request{C: strings.ToLower(args[0]), P: args[1:]}, nil
}

func (codec *Codec) parseArgument(rd *mut.Conn) (string, error) {
	line, err := rd.ReadString('\n')

	if err != nil {
		return "", io.EOF
	} else if len(line) < 3 {
		return "", io.EOF
	} else if line[0] != codeStrLen {
		return "", ErrInvalidRequest
	}

	blen, err := strconv.Atoi(line[1 : len(line)-2])
	if err != nil {
		return "", ErrInvalidRequest
	}

	buf := make([]byte, blen+2)
	if _, err := rd.ReadRaw(buf); err != nil {
		return "", io.EOF
	}

	return string(buf[:blen]), nil
}
