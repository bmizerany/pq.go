package proto

import (
	"fmt"
)

func encodeParam(param interface{}) (int32, string) {
	var s string
	switch param.(type) {
	default:
		panic(fmt.Sprintf("unknown type for %T", param))
	case int, uint8, uint16, uint32, uint64, int8, int16, int32, int64:
		s = fmt.Sprintf("%d", param)
	case string, []byte:
		s = fmt.Sprintf("%s", param)
	case bool:
		s = fmt.Sprintf("%t", param)
	}

	return int32(len(s)), s
}
