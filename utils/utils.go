package utils

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/copier"
	jsonIter "github.com/json-iterator/go"
	"github.com/satori/go.uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"path"
	"strings"
)

var (
	json = jsonIter.ConfigCompatibleWithStandardLibrary
)

// Copy repo: https://github.com/jinzhu/copier .
func Copy(from, to interface{}) error {
	return copier.Copy(to, from)
}

// CopyWithOption .
func CopyWithOption(from, to interface{}, option copier.Option) error {
	return copier.CopyWithOption(to, from, option)
}

// StructToJson .
func StructToJson(data interface{}) (string, error) {
	result, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// JsonToStruct .
func JsonToStruct(str string, data interface{}) error {
	if err := json.UnmarshalFromString(str, data); err != nil {
		return err
	}

	return nil
}

// MarshalEscapeHTML .
// @Param escapeHTML bool 是否转义 HTML 编码 [<、>、&]
func MarshalEscapeHTML(data interface{}, escapeHTML bool) ([]byte, error) {
	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(escapeHTML)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}

	return bf.Bytes(), nil
}

// Marshal .
func Marshal(data interface{}) ([]byte, error) {
	if m, ok := data.(proto.Message); ok {
		return protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(m)
	}

	result, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UnMarshal .
func UnMarshal(data []byte, v interface{}) error {
	if m, ok := v.(proto.Message); ok {
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, m)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}

// UNMarshal .
// Deprecated: use Unmarshal instead.
func UNMarshal(str string, data interface{}) error {
	if err := json.UnmarshalFromString(str, data); err != nil {
		return err
	}

	return nil
}

// JsonToStructFormByte .
func JsonToStructFormByte(str []byte, data interface{}) error {
	if err := json.Unmarshal(str, data); err != nil {
		return err
	}

	return nil
}

// GetUUID .
func GetUUID() string {
	uid := uuid.NewV4()
	return uid.String()
}

// GetPureUUID .
func GetPureUUID() string {
	return strings.Replace(GetUUID(), "-", "", -1)
}

// AppendPath .
func AppendPath(paths ...string) string {
	return path.Join(paths...)
}

// ByteCountSI 字节单位转换
func ByteCountSI(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func RemoveDuplicates(keys []interface{}) []interface{} {
	// we need only to check the map's keys, so we use the empty struct as values
	// since it consumes 0 bytes of memory.
	processed := make(map[interface{}]struct{})
	uniq := make([]interface{}, 0)
	for _, key := range keys {
		// if the user ID has been processed already, we skip it
		if _, ok := processed[key]; ok {
			continue
		}
		// append a unique user ID to the resulting slice.
		uniq = append(uniq, key)
		// mark the user ID as existing.
		processed[key] = struct{}{}
	}

	return uniq
}
