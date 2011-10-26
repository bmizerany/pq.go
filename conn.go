package pq

import (
	"io"
	"os"
	"fmt"
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
	*buffer
	wc  io.ReadWriteCloser
	scr *scanner
}

func New(rwc io.ReadWriteCloser) *Conn {
	cn := &Conn{
		wc:  rwc,
		scr: scan(rwc),
		buffer: newBuffer(),
	}

	return cn
}

func (cn *Conn) Startup(params Values) os.Error {
	cn.setType(0)
	cn.writeInt32(ProtoVersion)
	for k, v := range params {
		cn.writeString(k)
		cn.writeString(v)
	}
	cn.writeString("")

	err := cn.flush()
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

func (cn *Conn) flush() os.Error {
	cn.setLength()
	_, err := cn.wc.Write(cn.bytes())
	cn.reset()
	return err
}
