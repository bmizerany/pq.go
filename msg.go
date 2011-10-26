package pq

import (
	"fmt"
	"os"
)

type header struct {
	Type   byte
	Length int32
}

type msg struct {
	header
	body []byte
	err os.Error

	status int
}

func (m *msg) parse() os.Error {
	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown server response %c", m.Type)
	}

	panic("not reached")
}
