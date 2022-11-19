package ueditor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrFileAsDir       = errors.New("file used as dir")
	ErrFileMissing     = errors.New("target file not found")
	ErrFileMetaMissing = errors.New("file meta not found")
)

// specially for editor config marshal
func lowerCamalMarshal(i any) []byte {
	tp := reflect.TypeOf(i)
	switch tp.Kind() {
	case reflect.Pointer:
		return lowerCamalMarshal(reflect.ValueOf(i).Elem().Interface())
	case reflect.Struct:
		val := reflect.ValueOf(i)
		totalFields := tp.NumField()
		tmp := make([]string, totalFields)
		for i := 0; i < totalFields; i++ {
			name := []byte(tp.Field(i).Name)
			if name[0] >= 'A' && name[0] <= 'Z' {
				name[0] += 'a' - 'A'
			}
			tmp[i] = fmt.Sprintf("\"%s\":%s", string(name),
				lowerCamalMarshal(val.Field(i).Interface()))
		}
		return []byte("{" + strings.Join(tmp, ",") + "}")
	case reflect.Slice:
		val := reflect.ValueOf(i)
		size := val.Len()
		tmp := make([]string, size)
		for i := 0; i < size; i++ {
			tmp[i] = string(lowerCamalMarshal(val.Index(i).Interface()))
		}
		return []byte("[" + strings.Join(tmp, ",") + "]")
	default:
		r, e := json.Marshal(i)
		if e != nil {
			panic(e)
		}
		return r
	}
}

// one level apply for struct only
func applyDefault(v any) {
	tp := reflect.TypeOf(v)
	if tp.Kind() != reflect.Pointer {
		panic("should be pointer of the struct")
	}
	val := reflect.ValueOf(v).Elem()
	totalFields := tp.Elem().NumField()
	for i := 0; i < totalFields; i++ {
		v := val.Field(i)
		tag := tp.Elem().Field(i).Tag.Get("default")
		if v.CanSet() && tag != "" && isEmptyValue(v) {
			switch v.Type().Kind() {
			case reflect.String:
				v.SetString(tag)
			case reflect.Bool:
				parsed, e := strconv.ParseBool(tag)
				if e != nil {
					panic(e)
				}
				v.SetBool(parsed)
			case reflect.Int:
				parsed, e := strconv.Atoi(tag)
				if e != nil {
					panic(e)
				}
				v.SetInt(int64(parsed))
			case reflect.Float64:
				parsed, e := strconv.ParseFloat(tag, 64)
				if e != nil {
					panic(e)
				}
				v.SetFloat(parsed)
			case reflect.Slice:
				if v.Type().Elem().Kind() != reflect.String {
					panic("for now, only string slice is supported")
				}
				v.Set(reflect.ValueOf(strings.Split(tag, "|")))
			default:
				panic("yet-implemented default value")
			}
		}
	}
}

func fileExist(path string) bool {
	_, e := os.Stat(path)
	if e == nil {
		return true
	}
	return !os.IsNotExist(e)
}

func dirExist(dir string) (bool, error) {
	stat, e := os.Stat(dir)
	if e == nil {
		if stat.IsDir() {
			return true, nil
		}
		return false, ErrFileAsDir
	}
	return false, nil
}

func saveFileContent(path string, content []byte) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write(content)
	return e
}

func isFullAlpha(s string) bool {
	for _, char := range s {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') {
			return false
		}
	}
	return false
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}
