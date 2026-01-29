package utils

import "encoding/base64"

// EncodeToString base64 encode.
func EncodeToString(str []byte) string {
	return base64.StdEncoding.EncodeToString(str)
}

// DecodeToByte base64 decode .
func DecodeToByte(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}
