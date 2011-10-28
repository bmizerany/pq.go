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
			cn.Pid = m.Secret
		case 'Z':
			return cn, nil
		}
	}

	panic("not reached")
}

func (cn *Conn) Prepare(query string) (driver.Stmt, os.Error) {
	name := query //TODO: use something unique and smaller

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
	return nil
}

func (stmt *Stmt) NumInput() int {
	return len(stmt.params)
}

func (stmt *Stmt) Exec(args []interface{}) (driver.Result, os.Error) {
	return nil, nil
}

func (stmt *Stmt) Query(args []interface{}) (driver.Rows, os.Error) {
	return nil, nil
}

func notExpected(c byte) {
	panic(fmt.Sprintf("pq: unexpected response from server (%c)", c))
}
