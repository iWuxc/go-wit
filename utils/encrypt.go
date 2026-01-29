package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"io"
	"io/ioutil"
)

// HashMd5ForString .
func HashMd5ForString(str string) string {
	return hex.EncodeToString(HashMd5ForStringToByte(str))
}

// HashMd5ForStringToByte .
func HashMd5ForStringToByte(str string) []byte {
	return HashMd5ForByteToByte([]byte(str))
}

// HashMd5ForByte .
func HashMd5ForByte(b []byte) string {
	return hex.EncodeToString(HashMd5ForByteToByte(b))
}

// HashMd5ForByteToByte .
func HashMd5ForByteToByte(b []byte) []byte {
	h := md5.New()
	h.Write(b)
	return h.Sum(nil)
}

// HashMd5ForReader .
func HashMd5ForReader(reader io.Reader) string {
	h := md5.New()
	r := bufio.NewReader(reader)

	if _, err := io.Copy(h, r); err != nil {
		return ""
	}

	return hex.EncodeToString(h.Sum(nil))
}

// HashMd5ForFile .
func HashMd5ForFile(file string) string {
	data, _ := ioutil.ReadFile(file)

	return HashMd5ForString(string(data))
}

// HashSha256ForString .
func HashSha256ForString(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// HashSha256ForFile .
func HashSha256ForFile(file string) string {
	data, _ := ioutil.ReadFile(file)

	return HashSha256ForString(string(data))
}

// EncodePassword 对密码加密.
func EncodePassword(passwd string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash[:])
}

// ComparePassword 比较密码 .
func ComparePassword(passwd, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(passwd))
	return err == nil
}
