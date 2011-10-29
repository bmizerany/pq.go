package pq

import (
	"net"
	"exp/sql"
	"exp/sql/driver"
	"fmt"
	"github.com/bmizerany/pq.go/proto"
	"io"
	"os"
	"url"
)

type Driver struct{}

func (dr *Driver) Open(name string) (driver.Conn, os.Error) {
	u, err := url.Parse(name)
	if err != nil {
		return nil, err
	}

	nc, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, err
	}

	// TODO: use pass
	user, _, err := url.UnescapeUserinfo(u.RawUserinfo)
	if err != nil {
		return nil, err
	}

	params := make(proto.Values)
	params.Set("user", user)

	return New(nc, params)
}

var pgDriver = &Driver{}

func init() {
	sql.Register("postgres", pgDriver)
}

type Conn struct {
	Settings proto.Values
	Pid      int
	Secret   int
	Status   byte
	Notifies <-chan *proto.Notify

	rwc io.ReadWriteCloser
	p   *proto.Conn
	err os.Error
}

func New(rwc io.ReadWriteCloser, params proto.Values) (*Conn, os.Error) {
	notifies := make(chan *proto.Notify, 5) // 5 should be enough to prevent simple blocking

	cn := &Conn{
		Notifies: notifies,
		Settings: make(proto.Values),
		p:        proto.New(rwc, notifies),
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
	err    os.Error
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
		case 'n':
			// no data
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

func (stmt *Stmt) Close() (err os.Error) {
	err = stmt.p.Close(proto.Statement, stmt.Name)
	if err != nil {
		return err
	}

	err = stmt.p.Sync()
	if err != nil {
		return err
	}

	var done bool
	for {
		m, err := stmt.p.Next()
		if err != nil {
			return err
		}
		if m.Err != nil {
			return m.Err
		}

		if m.Type == '3' {
			done = true
		}

		if done && m.Type == 'Z' {
			return nil
		}
	}

	panic("not reached")
}

func (stmt *Stmt) NumInput() int {
	return len(stmt.params)
}

func (stmt *Stmt) Exec(args []interface{}) (driver.Result, os.Error) {
	// NOTE: should return []drive.Result, because a PS can have more
	// than one statement and recv more than one tag.
	rows, err := stmt.Query(args)
	if err != nil {
		return nil, err
	}

	for rows.Next(nil) != os.EOF {}

	// TODO: use the tag given by CommandComplete
	return driver.RowsAffected(0), nil
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
	c int
}

func (r *Rows) Close() (err os.Error) {
	// Drain the remaining rows
	for err == nil { err = r.Next(nil) }

	if err == os.EOF {
		return nil
	}

	return
}

func (r *Rows) Complete() int {
	return r.c
}

func (r *Rows) Columns() []string {
	return r.names
}

func (r *Rows) Next(dest []interface{}) (err os.Error) {
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

func (cn *Conn) Begin() (driver.Tx, os.Error) { panic("todo") }
func (cn *Conn) Close() os.Error { panic("todo") }

func notExpected(c byte) {
	panic(fmt.Sprintf("pq: unexpected response from server (%c)", c))
}
