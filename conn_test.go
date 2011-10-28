package pq

import (
	"github.com/bmizerany/assert"
	"net"
	"testing"
	"os"
)

func TestConnPrepare(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)

	cn, err := New(nc, map[string]string{"user": os.Getenv("USER")})
	assert.Equalf(t, nil, err, "%v", err)

	_, err = cn.Prepare("SELECT length($1) AS ZOMG! AN ERR")
	assert.Equalf(t, nil, err, "%v", err)
}
