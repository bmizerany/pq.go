package pq

import (
	"exp/sql/driver"
	"github.com/bmizerany/pq.go/proto"
	"io"
	"os"
)

type Conn struct {
	Params proto.Values
	rwc    io.ReadWriteCloser
}

func New(rwc io.ReadWriteCloser) *Conn {
	return nil
}

func (cn *Conn) Prepare(query string) (driver.Stmt, os.Error) {
	return nil, nil
}
