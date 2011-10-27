package pq

import (
	"encoding/binary"
	"fmt"
	"github.com/bmizerany/pq.go/buffer"
	"io"
	"os"
)

const ProtoVersion = int32(196608)

type Values map[string]string

func (vs Values) Get(k string) string {
	if v, ok := vs[k]; ok {
		return v
	}
	return ""
}

func (vs Values) Set(k, v string) {
	vs[k] = v
}

func (vs Values) Del(k string) {
	vs[k] = "", false
}

type Conn struct {
	b   *buffer.Buffer
	scr *scanner
	wc  io.ReadWriteCloser
}

func New(rwc io.ReadWriteCloser) *Conn {
	cn := &Conn{
		b:   buffer.New(nil),
		wc:  rwc,
		scr: scan(rwc),
	}

	return cn
}

func (cn *Conn) Startup(params Values) os.Error {
	cn.b.WriteInt32(ProtoVersion)
	for k, v := range params {
		cn.b.WriteCString(k)
		cn.b.WriteCString(v)
	}
	cn.b.WriteCString("")

	err := cn.flush(0)
	if err != nil {
		return err
	}

	m, err := cn.nextMsg()
	if err != nil {
		return err
	}

	err = m.parse()
	if err != nil {
		return err
	}

	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown startup response (%c)", m.Type)
	case 'E':
		return m.err
	case 'R':
		switch m.status {
		default:
			return fmt.Errorf("pq: unknown authentication type (%d)", m.status)
		case 0:
			return nil
		}
	}

	panic("not reached")
}

func (cn *Conn) nextMsg() (*msg, os.Error) {
	m, ok := <-cn.scr.msgs
	if !ok {
		return nil, cn.scr.err
	}
	return m, nil
}

func (cn *Conn) flush(t byte) os.Error {
	if t > 0 {
		err := binary.Write(cn.wc, binary.BigEndian, t)
		if err != nil {
			return err
		}
	}

	l := int32(cn.b.Len()) + sizeOfInt32
	err := binary.Write(cn.wc, binary.BigEndian, l)
	if err != nil {
		return err
	}

	_, err = cn.b.WriteTo(cn.wc)
	if err != nil {
		return err
	}

	return err
}
