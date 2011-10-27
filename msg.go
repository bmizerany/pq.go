package pq

import (
	"fmt"
	buffer "github.com/bmizerany/pq.go/buffer"
	"os"
)

type header struct {
	Type byte
	Length int32
}

type msg struct {
	header
	*buffer.Buffer
	err os.Error

	auth int

	status byte

	key, val string
	pid, secret int
}

func (m *msg) parse() os.Error {
	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown server response (%c)", m.Type)
	case 'R':
		m.auth = int(m.ReadInt32())
	case 'S':
		m.key = m.ReadCString()
		m.val = m.ReadCString()
	case 'K':
		m.pid = int(m.ReadInt32())
		m.secret = int(m.ReadInt32())
	case 'Z':
		m.status = m.ReadByte()
	case '1':
		// Nothing to read
	}

	return nil
}
