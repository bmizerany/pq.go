package pq

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"testing"
)

type readWriteLogger struct {
	io.ReadWriteCloser
}

func (rwl *readWriteLogger) Write(p []byte) (int, error) {
	fmt.Printf("%q\n", p)
	return rwl.ReadWriteCloser.Write(p)
}

func (rwl *readWriteLogger) Read(p []byte) (int, error) {
	defer fmt.Printf("%q\n", p)
	return rwl.ReadWriteCloser.Read(p)
}

func TestSimple(t *testing.T) {
	db, err := sql.Open("postgres", "sslmode=disable user="+os.Getenv("USER"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r, err := db.Query("SELECT 1")
	if err != nil {
		t.Fatal(err)
	}

	if !r.Next() {
		if r.Err() != nil {
			t.Fatal(r.Err())
		}
		t.Fatal("row expected")
	}

	var i int
	err = r.Scan(&i)
	if err != nil {
		t.Fatal(err)
	}

	if i != 1 {
		t.Fatal("expected i to be 1")
	}
}
