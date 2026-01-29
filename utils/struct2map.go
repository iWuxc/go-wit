package utils

import (
	"reflect"
	"sync"
)

var sw sync.RWMutex

func StructConvertMapWithJsonValue(src interface{}, expect map[string]interface{}, isDebug ...bool) map[string]string {
	var (
		m1 []string
		m2 = make(map[string]map[string]string)
	)

	if len(isDebug) == 0 {
		isDebug = append(isDebug, false)
	}

	for s, i := range expect {
		m1 = append(m1, s)
		m2[s] = StructConvertMapForJson(i, isDebug[0])
	}

	m := StructConvertMapForJson(src, isDebug[0], m1...)

	for str, val := range m2 {
		j, _ := json.Marshal(val)
		m[str] = string(j)
	}

	return m
}

func StructConvertMapForJson(src interface{}, isDebug bool, except ...string) map[string]string {
	m := make(map[string]string)
	if isDebug {
		StructConvertMapWithRemark(src, m, "json", except...)
	} else {
		StructConvertMap(src, m, "json", except...)
	}
	return m
}

func StructConvertMapForQuery(src interface{}, isDebug bool, except ...string) map[string]string {
	m := make(map[string]string)
	if isDebug {
		StructConvertMapWithRemark(src, m, "query", except...)
	} else {
		StructConvertMap(src, m, "query", except...)
	}
	return m
}

// StructConvertMap .
func StructConvertMap(src interface{}, dst map[string]string, tagName string, except ...string) {
	t := reflect.TypeOf(src)
	v := reflect.ValueOf(src)

	switch t.Kind().String() {
	case "struct":
	R:
		for i := 0; i < t.NumField(); i++ {
			tg := t.Field(i).Tag.Get(tagName)
			// except
			for _, e := range except {
				if e == tg {
					continue R
				}
			}

			if t.Field(i).Type.String() == "string" {
				sw.Lock()
				dst[tg] = v.Field(i).String()
				sw.Unlock()
				continue
			}

			StructConvertMap(v.Field(i).Interface(), dst, tagName, except...)
		}

	}
}

// StructConvertMapWithRemark .
func StructConvertMapWithRemark(src interface{}, dst map[string]string, tagName string, except ...string) {
	t := reflect.TypeOf(src)
	v := reflect.ValueOf(src)

	switch t.Kind().String() {
	case "struct":
	R:
		for i := 0; i < t.NumField(); i++ {
			tg := t.Field(i).Tag.Get(tagName)
			// except
			for _, e := range except {
				if e == tg {
					continue R
				}
			}

			if t.Field(i).Type.String() == "string" {
				sw.Lock()
				dst[tg] = t.Field(i).Tag.Get("remark")
				sw.Unlock()
				continue
			}

			StructConvertMap(v.Field(i).Interface(), dst, tagName, except...)
		}

	}
}

// VerifyStruct .
func VerifyStruct(src interface{}, fields []string) []string {
	var err []string
	t := reflect.TypeOf(src)
	v := reflect.ValueOf(src)

	switch t.Kind().String() {
	case "struct":
	R:
		for i := 0; i < t.NumField(); i++ {
			tg := t.Field(i).Tag.Get("json")
			if t.Field(i).Type.String() == "string" {
				for _, field := range fields {
					if tg == field && len(v.Field(i).String()) == 0 {
						err = append(err, tg+" not be empty")
						continue R
					}
				}
				continue
			}

			err = append(err, VerifyStruct(v.Field(i).Interface(), fields)...)
		}
	}

	return err
}
