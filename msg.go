package pq

import (
	"fmt"
)

type header struct {
	Type   byte
	Length int32
}

type msg struct {
	header
	body []byte
}

func (m *msg) decode() {
	switch m.Type {
	default:
		panic(fmt.Sprintf("pq: unknown server response %c", m.Type))
	}
}
