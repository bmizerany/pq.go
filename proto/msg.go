package proto

import (
	"fmt"
	"bytes"
)

const (
	AuthOk = iota
	_
	_
	AuthPlain
	_
	AuthMd5
)

const  (
	ErrorFieldSeverity = 'S'
	ErrorFieldCode = 'C'
	ErrorFieldMessage = 'M'
	ErrorFieldDetail = 'D'
	ErrorFieldHint = 'H'
	ErrorFieldPosition = 'P'
	ErrorFieldInternalPosition = 'p'
	ErrorFieldWhere = 'W'
	ErrorFieldFile = 'F'
	ErrorFieldLine = 'L'
	ErrorFieldRoutine = 'R'
)

type Header struct {
	Type   byte
	Length int32
}

type Msg struct {
	Header
	*Buffer
	Err         *Error
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

type Error struct {
	Fields map[byte]string
}

func (err *Error) Error() string {

	b := bytes.NewBufferString("pq: ")

	for fieldName, value := range err.Fields {
		fmt.Fprintf(b, "%s:%s,", readableFieldNames[fieldName], value)
	}

	return b.String()
}

func (m *Msg) parse() error {
	switch m.Type {
	default:
		return fmt.Errorf("pq: unknown server response (%c)", m.Type)
	case 'E':
		return m.ParseError()
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
			if n := int(m.ReadInt32()); n >= 0 {
				m.Cols[i] = make([]byte, n)
				m.Read(m.Cols[i])
			} else {
				m.Cols[i] = nil
			}
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

func (m *Msg) ParseError() error {
	fields := make(map[byte]string)

	var err error
	var status byte

	for status, err = m.Buffer.Buffer.ReadByte(); status != 0 && err == nil; status, err = m.Buffer.Buffer.ReadByte() {
		message := m.ReadCString()
		fields[status] = message
	}

	if err != nil {
		return err
	}

	m.Err = &Error{fields}

	return nil
}
var readableFieldNames = map[byte]string{
	ErrorFieldSeverity: "Severity",
	ErrorFieldCode: "Code",
	ErrorFieldMessage: "Message",
	ErrorFieldDetail: "Detail",
	ErrorFieldHint: "Hint",
	ErrorFieldPosition: "Position",
	ErrorFieldInternalPosition: "InternalPosition",
	ErrorFieldWhere: "Where",
	ErrorFieldFile: "File",
	ErrorFieldLine: "Line",
	ErrorFieldRoutine: "Routine",
}
