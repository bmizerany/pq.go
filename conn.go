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

	rwc io.ReadWriteCloser
	p   *proto.Conn
	err os.Error
}

func New(rwc io.ReadWriteCloser, params proto.Values) (*Conn, os.Error) {
	cn := &Conn{
		Settings: make(proto.Values),
		p:        proto.New(rwc),
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
			cn.Secret = m.Secret
		case 'Z':
			return cn, nil
		}
	}

	panic("not reached")
}

func (cn *Conn) Prepare(query string) (driver.Stmt, os.Error) {
	name := "" //TODO: support named queries

	stmt := &Stmt{
		Name:  name,
		query: query,
		p:     cn.p,
	}

	err := stmt.Parse()
	if err != nil {
		return nil, err
	}

	err = stmt.Describe()
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

type Stmt struct {
	Name string

	query  string
	p      *proto.Conn
	params []int
	names  []string
}

func (stmt *Stmt) Parse() os.Error {
	err := stmt.p.Parse(stmt.Name, stmt.query)
	if err != nil {
		return err
	}

	err = stmt.p.Sync()
	if err != nil {
		return err
	}

	for {
		m, err := stmt.p.Next()
		if err != nil {
			return err
		}
		if m.Err != nil {
			return m.Err
		}

		switch m.Type {
		default:
			notExpected(m.Type)
		case '1':
			// ignore
		case 'Z':
			return nil
		}
	}

	panic("not reached")
}

func (stmt *Stmt) Describe() os.Error {
	err := stmt.p.Describe(proto.Statement, stmt.Name)
	if err != nil {
		return err
	}

	err = stmt.p.Sync()
	if err != nil {
		return err
	}

	for {
		m, err := stmt.p.Next()
		if err != nil {
			return err
		}
		if m.Err != nil {
			return m.Err
		}

		switch m.Type {
		default:
			notExpected(m.Type)
		case 't':
			stmt.params = m.Params
		case 'T':
			stmt.names = m.ColNames
		case 'Z':
			return nil
		}
	}

	panic("not reached")
}

func (stmt *Stmt) Close() os.Error {
	panic("todo")
}

func (stmt *Stmt) NumInput() int {
	return len(stmt.params)
}

func (stmt *Stmt) Exec(args []interface{}) (driver.Result, os.Error) {
	panic("todo")
}

func (stmt *Stmt) Query(args []interface{}) (driver.Rows, os.Error) {
	// For now, we'll just say they're strings
	sargs := encodeParams(args)

	err := stmt.p.Bind(stmt.Name, stmt.Name, sargs...)
	if err != nil {
		return nil, err
	}

	err = stmt.p.Execute(stmt.Name, 0)
	if err != nil {
		return nil, err
	}

	err = stmt.p.Sync()
	if err != nil {
		return nil, err
	}

	for {
		m, err := stmt.p.Next()
		if err != nil {
			return nil, err
		}
		if m.Err != nil {
			return nil, m.Err
		}

		switch m.Type {
		default:
			notExpected(m.Type)
		case '2':
			rows := &Rows{
				p:     stmt.p,
				names: stmt.names,
			}
			return rows, nil
		}
	}

	panic("not reached")
}

type Rows struct {
	p *proto.Conn
	names []string
	err os.Error
	c int
}

func (r *Rows) Close() os.Error {
	return nil
}

func (r *Rows) Complete() int {
	return r.c
}

func (r *Rows) Columns() []string {
	return r.names
}

func (r *Rows) Next(dest []interface{}) (err os.Error) {
	if r.err != nil {
		return r.err
	}

	defer func() {
		r.err = err
	}()

	var m *proto.Msg
	for {
		m, err = r.p.Next()
		if err != nil {
			return err
		}
		if m.Err != nil {
			return m.Err
		}

		switch m.Type {
		default:
			notExpected(m.Type)
		case 'D':
			for i := 0; i < len(dest); i++ {
				dest[i] = string(m.Cols[i])
			}
			return nil
		case 'C':
			r.c++
		case 'Z':
			return os.EOF
		}
	}

	panic("not reached")
}

func notExpected(c byte) {
	panic(fmt.Sprintf("pq: unexpected response from server (%c)", c))
}
