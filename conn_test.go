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
	db, err := sql.Open("postgres", "sslmode=disable user=pqgotest password=foo")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	r, err := db.Query("SELECT 1, 2")
	if err != nil {
		t.Fatal(err)
	}

	if !r.Next() {
		if r.Err() != nil {
			t.Fatal(r.Err())
		}
		t.Fatal("row expected")
	}

	var x, y int
	err = r.Scan(&x, &y)
	if err != nil {
		t.Fatal(err)
	}

	if x != 1 {
		t.Fatal("expected x to be 1")
	}

	if y != 2 {
		t.Fatal("expected y to be 2")
	}
}

func TestMultipleQueries(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost user=pqgotest password=foo sslmode=disable")
	if err != nil {
		t.Fatalf("unable to open database connection: %v", err)
	}

	for i := 0; i < 5; i++ {
		var n int
		err = db.QueryRow("SELECT 1").Scan(&n)
		switch {
		case err != nil:
			t.Fatalf("%s: at %d", err, n)
		case n != 1:
			t.Fatalf("expected 1 at %d", n)
		}
	}
}

func TestError(t *testing.T) {
	db, err := sql.Open("postgres", "host=localhost user=pqgotest password=foo sslmode=disable")
	if err != nil {
		t.Fatalf("unable to open database connection: %v", err)
	}

	_, err = db.Query("SELECT holla!")
	if err == nil {
		t.Fatal("expected error")
	}

	if _, ok := err.(*ServerError); !ok {
		t.Fatal("expected *ServerError")
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
