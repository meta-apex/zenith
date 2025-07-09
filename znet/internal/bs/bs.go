package bs

import (
	"unsafe"
)

// BytesToString converts byte slice to a string without any memory allocation.
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// StringToBytes converts string to a byte slice without any memory allocation.
func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
