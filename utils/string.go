package utils

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// ParseInt .
func ParseInt(s string) int {
	if s == "" {
		return 0
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

// ParseBool .
func ParseBool(s string) bool {
	i := ParseInt(s)
	return i != 0
}

// CamelToSnakeCase . xxYy => xx_yy   XxYy => xx_yy
func CamelToSnakeCase(camel string) string {
	var buf bytes.Buffer
	for _, c := range camel {
		if 'A' <= c && c <= 'Z' {
			// just convert [A-Z] to _[a-z]
			if buf.Len() > 0 {
				buf.WriteRune('_')
			}
			buf.WriteRune(c - 'A' + 'a')
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

// ReadStringFromPath .
func ReadStringFromPath(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(data)
}

// WriteStringToPath .
func WriteStringToPath(s string, path string) {
	err := os.WriteFile(path, []byte(s), 0644)
	if err != nil {
		panic(err)
	}
}

// ReadBytesFromPath .
func ReadBytesFromPath(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return data
}

// WriteBytesToPath .
func WriteBytesToPath(b []byte, path string) {
	err := os.WriteFile(path, b, 0644)
	if err != nil {
		panic(err)
	}
}

// SnakeString XxYy to xx_yy
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

// IsChinese .
func IsChinese(str string) bool {
	var flag bool
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			flag = true
			break
		}
	}
	return flag
}

// ToString .
func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case nil:
		return ""
	case map[string]string, map[string]interface{}:
		d, _ := Marshal(v)
		return string(d)
	default:
		return fmt.Sprintf("%v", v)
	}
}
