package pq

type buffer struct {
	b []byte
	pos int
}

func newBuffer() *buffer {
	// return a buffer that starts after the header
	b := &buffer{b: make([]byte, 1024)}
	b.reset()
	return b
}

func (b *buffer) setType(c byte) {
	b.b[0] = c
}

func (b *buffer) setLength() {
	length := b.pos
	b.pos = 1
	b.writeInt32(int32(length-1)) // don't include Type in length
	b.pos = length
}

func (b *buffer) writeByte(c byte) {
	b.b[b.pos] = c
	b.pos += 1
}

func (b *buffer) writeString(v string) {
	for _, c := range v {
		b.writeByte(byte(c))
	}
	b.writeByte(0)
}

func (b *buffer) writeInt16(v int16) {
	b.writeByte(byte(v >> 8))
	b.writeByte(byte(v))
}

func (b *buffer) writeInt32(v int32) {
	b.writeByte(byte(v >> 24))
	b.writeByte(byte(v >> 16))
	b.writeByte(byte(v >> 8))
	b.writeByte(byte(v))
}

func (b *buffer) bytes() []byte {
	if b.b[0] == 0 {
		return b.b[1:b.pos]
	}
	return b.b[:b.pos]
}

func (b *buffer) reset() {
	b.pos = 5
	b.setType(0)
	b.setLength()
}
