package pq

import (
	"bytes"
	"encoding/binary"
	"github.com/bmizerany/pq.go/buffer"
	"io"
	"os"
)

const sizeOfInt32 = int32(32 / 8)

type scanner struct {
	r    io.Reader
	msgs <-chan *msg
	err  os.Error
}

func scan(r io.Reader) *scanner {
	msgs := make(chan *msg)
	s := &scanner{r: r, msgs: msgs}

	go s.run(msgs)

	return s
}

func (s *scanner) run(msgs chan<- *msg) {
	var err os.Error
	defer func() {
		s.err = err
		close(msgs)
	}()

	for {
		m := new(msg)

		err = binary.Read(s.r, binary.BigEndian, &m.header)
		if err != nil {
			return
		}
		m.Length -= sizeOfInt32

		b := make([]byte, m.Length)
		_, err = io.ReadFull(s.r, b)
		if err != nil {
			return
		}

		m.Buffer = &buffer.Buffer{bytes.NewBuffer(b)}

		msgs <- m
	}
}
