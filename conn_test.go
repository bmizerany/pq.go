package pq

import (
	"github.com/bmizerany/assert"
	"net"
	"os"
	"testing"
)

func TestConnStartupErr(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)
	defer nc.Close()

	cn := New(nc)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Startup(Values{})
	assert.NotEqual(t, nil, err)
}

func TestConnStartup(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)
	defer nc.Close()

	cn := New(nc)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Startup(Values{"user": os.Getenv("USER")})
	assert.Equalf(t, nil, err, "%v", err)
}

func TestConnQuery(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)
	defer nc.Close()

	cn := New(nc)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Startup(Values{"user": os.Getenv("USER")})
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Parse("test", "SELECT length($1) AS foo")
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Bind("test", "test", "testing")
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Execute("test", 0)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Recv()
	assert.Equalf(t, nil, err, "%v", err)

	m, err := cn.Next()
	assert.Equalf(t, nil, err, "%v", err)
	assert.Equalf(t, byte('D'), m.Type, "%c", m.Type)

	err = cn.Complete()
	assert.Equalf(t, nil, err, "%v", err)

	// Query 2

	err = cn.Bind("test", "test", "foobar")
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Execute("test", 0)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Recv()
	assert.Equalf(t, nil, err, "%v", err)

	m, err = cn.Next()
	assert.Equalf(t, nil, err, "%v", err)
	assert.Equalf(t, byte('D'), m.Type, "%c", m.Type)

	err = cn.Complete()
	assert.Equalf(t, nil, err, "%v", err)
}

func TestConnErr(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)
	defer nc.Close()

	cn := New(nc)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Startup(Values{"user": os.Getenv("USER")})
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Parse("test", "SELECT length($1) ZOMG! ERROR")
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Bind("test", "test", "testing")
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Execute("test", 0)
	assert.Equalf(t, nil, err, "%v", err)

	// Errors don't come until we sync
	err = cn.Recv()
	assert.NotEqual(t, nil, err)
}
