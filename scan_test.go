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
	got := <-s.msgs
	assert.Equal(t, header{'X', 8}, got.header)
	assert.Equal(t, "testing\x00", got.String())

	_, ok := <-s.msgs
	assert.Equal(t, false, ok)
	assert.Equal(t, os.EOF, s.err)
}
