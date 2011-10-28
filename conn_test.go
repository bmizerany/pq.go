package pq

import (
	"github.com/bmizerany/assert"
	"net"
	"testing"
)

func TestConnPrepare(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)

	cn := New(nc)

	_, err = cn.Prepare("SELECT length($1) AS ZOMG! AN ERR")
	assert.Equalf(t, nil, err, "%v", err)
}
