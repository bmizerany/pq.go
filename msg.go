package pq

import (
	"fmt"
)

type header struct {
	Mark   byte
	Length int32
}

type msg struct {
	header
	body []byte
}

func (m *msg) decode() {
	switch m.Mark {
	default:
		panic(fmt.Sprintf("pq: unknown server response %c", m.Mark))
	}
}
