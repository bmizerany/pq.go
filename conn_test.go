package pq

import (
	"github.com/bmizerany/assert"
	"net"
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
