package proto

import (
	"fmt"
	"os"
)

const (
	AuthOk = iota
	_
	_
	AuthPlain
	_
	AuthMd5
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
	Salt        string
	Status      byte
	Key, Val    string
	Pid, Secret int
	Cols        [][]byte
	ColNames    []string
	Tag         string
	Params      []int
	From        string
	Payload     string
	Message     string
}

func (m *Msg) parse() os.Error {
	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown server response (%c)", m.Type)
	case 'E':
		m.Status = m.ReadByte()
		m.Err = fmt.Errorf("pq: (%c) %s", m.Status, m.String())
		return nil // avoid the check at the end
	case 'R':
		m.Auth = int(m.ReadInt32())
		switch m.Auth {
		default:
			return fmt.Errorf("pq: unknown authentication type (%d)", m.Auth)
		case AuthOk, AuthPlain:
		case AuthMd5:
			m.Salt = string(m.Next(4))
		}
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
	case 'T':
		m.ColNames = make([]string, int(m.ReadInt16()))
		for i := 0; i < len(m.ColNames); i++ {
			m.ColNames[i] = m.ReadCString()
			// ignore the rest
			m.ReadInt32()
			m.ReadInt16()
			m.ReadInt32()
			m.ReadInt16()
			m.ReadInt32()
			m.ReadInt16()
		}
	case 'n':
		// ignore
	case 'A':
		m.Pid = int(m.ReadInt32())
		m.From = m.ReadCString()
		m.Payload = m.ReadCString()
	case '3':
		// ignore
	case 'N':
		m.Status = m.ReadByte()
		m.Message = m.String()
		return nil // avoid len check
	}

	if m.Len() != 0 {
		return fmt.Errorf("pq: %d unread bytes left in msg", m.Len())
	}

	return nil
}
