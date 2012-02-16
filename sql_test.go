package pq

import (
	"database/sql"
	"fmt"
	"github.com/bmizerany/assert"
	"os"
	"testing"
)

var cs = fmt.Sprintf("postgres://%s:@localhost:5432", os.Getenv("USER"))

func TestSqlSimple(t *testing.T) {
	cn, err := sql.Open("postgres", cs)
	assert.Equalf(t, nil, err, "%v", err)

	rows, err := cn.Query("SELECT length($1) AS foo", "testing")
	assert.Equalf(t, nil, err, "%v", err)

	ok := rows.Next()
	assert.T(t, ok)

	var length int
	err = rows.Scan(&length)
	assert.Equalf(t, nil, err, "%v", err)

	assert.Equal(t, 7, length)
}

func TestBinary(t *testing.T) {
	cn, err := sql.Open("postgres", cs)
	_, err = cn.Exec("DROP TABLE foo")
	if err != nil {
		t.Log(err)
	}

	_, err = cn.Exec("CREATE TABLE foo (b bytea)")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cn.Exec("INSERT INTO foo (b) VALUES ($1)", []byte{0, 1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}

	var b []byte
	err = cn.QueryRow("SELECT b FROM foo LIMIT 1").Scan(&b)
	if err != nil {
		t.Fatal(err)
	}
}
