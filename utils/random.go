package utils

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

var (
	src = rand.NewSource(time.Now().UnixNano())

	stringWord     = []byte(fmt.Sprintf("%s%s", stringLowerWord, stringUpperWord))
	stringWordLen  = len(stringWord)
	stringUWord    = []byte(fmt.Sprintf("%s%s", stringDigit, stringUpperWord))
	stringUWordLen = len(stringUWord)
	letterRunes    = []byte(fmt.Sprintf("%s%s%s%s", stringDigit, stringLowerWord, stringMisc, stringUpperWord))
	letterRunesLen = len(letterRunes)
)

const (
	stringUpperWord    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	stringLowerWord    = "abcdefghijklmnopqrstuvwxyz"
	stringDigit        = "1234567890"
	stringMisc         = ".$#@&*_"
	stringUpperWordLen = len(stringUpperWord)

	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandomString .
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(letterRunesLen)]
	}
	return string(b)
}

// RandLower .
func RandLower(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(stringLowerWord) {
			b[i] = stringLowerWord[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// RandUpper .
func RandUpper(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = stringUpperWord[rand.Intn(stringUpperWordLen)]
	}
	return string(b)
}

// RandWord .
func RandWord(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = stringWord[rand.Intn(stringWordLen)]
	}
	return string(b)
}

// RandNumber .
func RandNumber(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(stringDigit) {
			b[i] = stringDigit[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// RandUpperStringNumber .
func RandUpperStringNumber(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = stringUWord[rand.Intn(stringUWordLen)]
	}
	for string(b[0]) == "0" {
		b[0] = stringUWord[rand.Intn(stringUWordLen)]
	}
	return string(b)
}
