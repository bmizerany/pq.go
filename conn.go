package pq

import (
	"exp/sql/driver"
	"fmt"
	"github.com/bmizerany/pq.go/proto"
	"io"
	"os"
)

type Conn struct {
	Settings proto.Values
	Pid      int
	Secret   int
	Status   byte

	rwc      io.ReadWriteCloser
	p        *proto.Conn
	err      os.Error
}

func New(rwc io.ReadWriteCloser, params proto.Values) (*Conn, os.Error) {
	cn := &Conn{
		Settings: make(proto.Values),
		p: proto.New(rwc),
	}

	err := cn.p.Startup(params)
	if err != nil {
		return nil, err
	}

	for {
		m, err := cn.p.Next()
		if err != nil {
			return nil, err
		}

		if m.Err != nil {
			return nil, m.Err
		}

		switch m.Type {
		default:
			notExpected(m.Type)
		case 'R':
			switch m.Auth {
			default:
				return nil, fmt.Errorf("pq: unknown authentication type (%d)", m.Status)
			case 0:
				continue
			}
		case 'S':
			cn.Settings.Set(m.Key, m.Val)
		case 'K':
			cn.Pid = m.Pid
			cn.Pid = m.Secret
		case 'Z':
			return cn, nil
		}
	}

	panic("not reached")
}

func (cn *Conn) Prepare(query string) (driver.Stmt, os.Error) {
	return nil, nil
}

func notExpected(c byte) {
	panic(fmt.Sprintf("pq: unexpected response from server (%c)", c))
}
