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

	status int
}

func (m *msg) parse() os.Error {
	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown server response (%c)", m.Type)
	case 'R':
		m.status = int(m.ReadInt32())
	}

	return nil
}
