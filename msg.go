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

	cols [][]byte

	tag string
}

func (m *msg) parse() os.Error {
	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown server response (%c)", m.Type)
	case 'E':
		m.err = fmt.Errorf("pq: %s", m.String())
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
	case '1', '2':
		// Nothing to read
	case 'D':
		m.cols = make([][]byte, int(m.ReadInt16()))
		for i := 0; i < len(m.cols); i++ {
			m.cols[i] = make([]byte, int(m.ReadInt32()))
			m.Read(m.cols[i])
		}
	case 'C':
		m.tag = m.ReadCString()
	}

	if m.Len() != 0 {
		return fmt.Errorf("pq: %d unread bytes left in msg", m.Len())
	}

	return nil
}
