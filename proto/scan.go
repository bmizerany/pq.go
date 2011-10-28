package proto

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

const sizeOfInt32 = int32(32 / 8)

type scanner struct {
	r    io.Reader
	msgs <-chan *Msg
	err  os.Error
}

func scan(r io.Reader) *scanner {
	msgs := make(chan *Msg)
	s := &scanner{r: r, msgs: msgs}

	go s.run(msgs)

	return s
}

func (s *scanner) run(msgs chan<- *Msg) {
	var err os.Error
	defer func() {
		s.err = err
		close(msgs)
	}()

	for {
		m := new(Msg)

		err = binary.Read(s.r, binary.BigEndian, &m.Header)
		if err != nil {
			return
		}
		m.Length -= sizeOfInt32

		b := make([]byte, m.Length)
		_, err = io.ReadFull(s.r, b)
		if err != nil {
			return
		}

		m.Buffer = &Buffer{bytes.NewBuffer(b)}

		msgs <- m
	}
}
