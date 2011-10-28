package pq

import (
	"github.com/bmizerany/assert"
	"net"
	"testing"
	"os"
)

func TestConnPrepareErr(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)

	cn, err := New(nc, map[string]string{"user": os.Getenv("USER")})
	assert.Equalf(t, nil, err, "%v", err)

	_, err = cn.Prepare("SELECT length($1) AS ZOMG! AN ERR")
	assert.NotEqual(t, nil, err)
}

func TestConnPrepare(t *testing.T) {
	nc, err := net.Dial("tcp", "localhost:5432")
	assert.Equalf(t, nil, err, "%v", err)

	cn, err := New(nc, map[string]string{"user": os.Getenv("USER")})
	assert.Equalf(t, nil, err, "%v", err)

	stmt, err := cn.Prepare("SELECT length($1) AS foo WHERE true = $2")
	assert.Equalf(t, nil, err, "%v", err)
	assert.Equal(t, 2, stmt.NumInput())

	rows, err := stmt.Query([]interface{}{"testing", true})
	assert.Equalf(t, nil, err, "%v", err)
	assert.Equal(t, []string{"foo"}, rows.Columns())

	dest := make([]interface{}, 1)
	err = rows.Next(dest)
	assert.Equalf(t, nil, err, "%v", err)
	assert.Equal(t, []interface{}{"7"}, dest)

	err = rows.Next(dest)
	assert.Equalf(t, os.EOF, err, "%v", err)
	err = rows.Next(dest)
	assert.Equalf(t, os.EOF, err, "%v", err)

	rows, err = stmt.Query([]interface{}{"testing", false})
	assert.Equalf(t, nil, err, "%v", err)
	assert.Equal(t, []string{"foo"}, rows.Columns())

	err = rows.Next(dest)
	assert.Equalf(t, os.EOF, err, "%v", err)
}
