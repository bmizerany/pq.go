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

	cn := New(nc)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Startup(Values{})
	assert.NotEqual(t, nil, err)
}

func TestConnStartup(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)

	cn := New(nc)
	assert.Equalf(t, nil, err, "%v", err)

	err = cn.Startup(Values{"user": os.Getenv("USER")})
	assert.Equalf(t, nil, err, "%v", err)
}
