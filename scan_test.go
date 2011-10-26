package pq

import (
	"bytes"
	"github.com/bmizerany/assert"
	"testing"
	"os"
)

func TestScanSimple(t *testing.T) {
	b := bytes.NewBufferString("X\x00\x00\x00\x0Ctesting\x00")
	s := scan(b)
	exp := &msg{header:header{'X', 12}, body:[]byte("testing\x00")}
	got := <-s.msgs
	assert.Equal(t, exp.header, got.header)
	assert.Equal(t, exp.body, got.body)

	_, ok := <-s.msgs
	assert.Equal(t, false, ok)
	assert.Equal(t, os.EOF, s.err)
}
