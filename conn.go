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
	Settings Values
	Pid int
	Secret int
	Status byte

	b   *buffer.Buffer
	scr *scanner
	wc  io.ReadWriteCloser
}

func New(rwc io.ReadWriteCloser) *Conn {
	cn := &Conn{
		Settings: make(Values),
		b:   buffer.New(nil),
		wc:  rwc,
		scr: scan(rwc),
	}

	return cn
}

func (cn *Conn) Next() (*Msg, os.Error) {
	m, ok := <-cn.scr.msgs
	if !ok {
		return nil, cn.scr.err
	}
	if err := m.parse(); err != nil {
		return nil, err
	}
	return m, nil
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

	for {
		m, err := cn.Next()
		if err != nil {
			return err
		}

		if m.Err != nil {
			return m.Err
		}

		switch m.Type {
		default:
			return fmt.Errorf("pq: unknown startup response (%c)", m.Type)
		case 'R':
			switch m.Auth {
			default:
				return fmt.Errorf("pq: unknown authentication type (%d)", m.Status)
			case 0:
				continue
			}
		case 'S':
			cn.Settings.Set(m.Key, m.Val)
		case 'K':
			cn.Pid = m.Pid
			cn.Pid = m.Secret
		case 'Z':
			return nil
		}
	}

	panic("not reached")
}

func (cn *Conn) Parse(name, query string) os.Error {
	cn.b.WriteCString(name)
	cn.b.WriteCString(query)
	cn.b.WriteInt16(0)
	return cn.flush('P')
}

func (cn *Conn) Bind(portal, stmt string, args ... string) os.Error {
	cn.b.WriteCString(portal)
	cn.b.WriteCString(stmt)

	// TODO: Use format codes; maybe?
	//       some thought needs to be put into the design of this.
	//       See (Bind) http://developer.postgresql.org/pgdocs/postgres/protocol-message-formats.html
	cn.b.WriteInt16(0)

	cn.b.WriteInt16(int16(len(args)))
	for _, arg := range args {
		cn.b.WriteInt32(int32(len(arg)))
		cn.b.WriteString(arg)
	}

	// TODO: Use result format codes; maybe?
	//       some thought needs to be put into the design of this.
	//       See (Bind) http://developer.postgresql.org/pgdocs/postgres/protocol-message-formats.html
	cn.b.WriteInt16(0)

	return cn.flush('B')
}

func (cn *Conn) Execute(name string, rows int) os.Error {
	cn.b.WriteCString(name)
	cn.b.WriteInt32(int32(rows))
	return cn.flush('E')
}

func (cn *Conn) Sync() os.Error {
	err := cn.flush('S')
	if err != nil {
		return err
	}
	return nil
}

func (cn *Conn) Recv() os.Error {
	err := cn.Sync()
	if err != nil {
		return err
	}

	m, err := cn.Next()
	if err != nil {
		return err
	}
	if m.Err != nil {
		err := cn.Ready()
		if err != nil {
			return err
		}
		return m.Err
	}

	if m.Type == '2' {
		return nil
	}

	if m.Type != '1' {
		notWanted('1', m.Type)
	}

	m, err = cn.Next()
	if err != nil {
		return err
	}

	if m.Type != '2' {
		notWanted('2', m.Type)
	}

	return nil
}

func (cn *Conn) Ready() os.Error {
	m, err := cn.Next()
	if err != nil {
		return err
	}

	if m.Type != 'Z' {
		notWanted('Z', m.Type)
	}

	return nil
}

func (cn *Conn) Complete() os.Error {
	m, err := cn.Next()
	if err != nil {
		return err
	}
	if m.Type != 'C' {
		notWanted('C', m.Type)
	}

	m, err = cn.Next()
	if err != nil {
		return err
	}
	if m.Type != 'Z' {
		notWanted('Z', m.Type)
	}
	return err
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

func (cn *Conn) Close() os.Error {
	return cn.wc.Close()
}

func notWanted(w, g byte) {
	panic(fmt.Sprintf("pq: response %c expected, but got %c", w, g))
}
