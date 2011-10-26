package pq

import (
	"testing"
	"github.com/bmizerany/assert"
)

func TestBufferSimple(t *testing.T) {
	buf := newBuffer()
	buf.setType('X')
	buf.writeString("testing")
	buf.writeInt16(1)
	buf.writeInt32(2)
	buf.setLength()
	assert.Equal(t, "X\x00\x00\x00\x12testing\x00\x00\x01\x00\x00\x00\x02", string(buf.bytes()))
}

func TestBufferType(t *testing.T) {
	buf := newBuffer()
	buf.setType('A')
	assert.Equal(t, "A\x00\x00\x00\x04", string(buf.bytes()))
	buf.setType(0)
	assert.Equal(t, "\x00\x00\x00\x04", string(buf.bytes()))
}

func TestBufferReset(t *testing.T) {
	buf := newBuffer()
	buf.setType('X')
	buf.writeString("testing")
	buf.reset()
	assert.Equal(t, "\x00\x00\x00\x04", string(buf.bytes()))
}
