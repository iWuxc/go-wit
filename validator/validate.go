package validator

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"strconv"
)

// ValidateErr .
func ValidateErr(err error) string {
	if _, ok := err.(validator.ValidationErrors); ok {
		return Translate(err.(validator.ValidationErrors))
	} else if _, ok1 := err.(*strconv.NumError); ok1 {
		return err.(*strconv.NumError).Num + " 类型不正确"
	} else if _, ok2 := err.(*json.UnmarshalTypeError); ok2 {
		return err.(*json.UnmarshalTypeError).Field + " 类型不正确"
	}
	return err.Error()
}
