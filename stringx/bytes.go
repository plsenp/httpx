package stringx

import "unsafe"

func ToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func FromBytes(data []byte) string {
	return unsafe.String(&data[0], len(data))
}
