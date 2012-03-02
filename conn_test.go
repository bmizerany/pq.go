package pq

import (
	"database/sql"
	"fmt"
	"io"
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
	db, err := sql.Open("postgres", "sslmode=require user=pqgotest password=foo")
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

func TestSimpleParseURL(t *testing.T) {
	expected := "host=hostname.remote"
	str, err := ParseURL("postgres://hostname.remote")
	if err != nil {
		t.Fatal(err)
	}

	if str != expected {
		t.Fatalf("unexpected result from ParseURL:\n+ %v\n- %v", str, expected)
	}
}

func TestFullParseURL(t *testing.T) {
	expected := "port=1234 host=hostname.remote user=username password=secret dbname=database"
	str, err := ParseURL("postgres://username:secret@hostname.remote:1234/database")
	if err != nil {
		t.Fatal(err)
	}

	if str != expected {
		t.Fatalf("unexpected result from ParseURL:\n+ %s\n- %s", str, expected)
	}
}

func TestInvalidProtocolParseURL(t *testing.T) {
	_, err := ParseURL("http://hostname.remote")
	switch err {
	case nil:
		t.Fatal("Expected an error from parsing invalid protocol")
	default:
		msg := "invalid connection protocol: http"
		if err.Error() != msg {
			t.Fatal("Unexpected error message:\n+ %s\n- %s", err.Error(), msg)
		}
	}
}
