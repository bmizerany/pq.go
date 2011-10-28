package proto

import (
	"fmt"
	"os"
)

type Header struct {
	Type   byte
	Length int32
}

type Msg struct {
	Header
	*Buffer
	Err         os.Error
	Auth        int
	Status      byte
	Key, Val    string
	Pid, Secret int
	Cols        [][]byte
	Tag         string
	Params      []int
}

func (m *Msg) parse() os.Error {
	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown server response (%c)", m.Type)
	case 'E':
		m.Err = fmt.Errorf("pq: %s", m.String())
		return nil // avoid the check at the end
	case 'R':
		m.Auth = int(m.ReadInt32())
	case 'S':
		m.Key = m.ReadCString()
		m.Val = m.ReadCString()
	case 'K':
		m.Pid = int(m.ReadInt32())
		m.Secret = int(m.ReadInt32())
	case 'Z':
		m.Status = m.ReadByte()
	case '1', '2':
		// Nothing to read
	case 'D':
		m.Cols = make([][]byte, int(m.ReadInt16()))
		for i := 0; i < len(m.Cols); i++ {
			m.Cols[i] = make([]byte, int(m.ReadInt32()))
			m.Read(m.Cols[i])
		}
	case 'C':
		m.Tag = m.ReadCString()
	case 't':
		m.Params = make([]int, int(m.ReadInt16()))
		for i := 0; i < len(m.Params); i++ {
			m.Params[i] = int(m.ReadInt32())
		}
	}

	if m.Len() != 0 {
		return fmt.Errorf("pq: %d unread bytes left in msg", m.Len())
	}

	return nil
}
