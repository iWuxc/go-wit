package utils

import (
	"bytes"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
)

// CheckContentType 检测文件编码 .
func CheckContentType(data []byte) string {
	_, name, _ := charset.DetermineEncoding(data, "")
	return name
}

// ConvGBKToUTF8 转换编码 GBK => UTF-8 .
func ConvGBKToUTF8(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// ConvUTF8ToGBK 转换编码 UTF-8 => GBK .
func ConvUTF8ToGBK(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}
