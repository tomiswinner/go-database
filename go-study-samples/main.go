package main

import (
	"fmt"
	"reflect"
	"strconv"
)

type MapStruct struct {
	Str string `map:"str"`
	StrPtr *string `map:"str"`
	Int int `map:"int"`
	IntPtr *int `map:"int"`
	Bool bool `map:"bool"`
	BoolPtr *bool `map:"bool"`
}

func main() {
	src := map[string]string{
		"str": "string data",
		"bool": "true",
		"int": "123",
	}

	var ms MapStruct
	Decode(&ms, src)
	fmt.Println("ms: ", ms)
}

func Decode(dst interface{}, src map[string]string) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr {
		// ポインタ出ない場合、値のコピーになり、元の値を変更できないのでエラー
		return fmt.Errorf("dst must be a pointer")
	}

	v = v.Elem() // ポインタの実体を取得
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("map")
		if tag == "" {
			continue
		}
		value, exists := src[tag]
		if !exists {
			continue
		}

		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int:
			if i, err := strconv.Atoi(value); err == nil {
				fieldValue.SetInt(int64(i))
			}
		case reflect.Bool:
			if b, err := strconv.ParseBool(value); err == nil {
				fieldValue.SetBool(b)
			}
		case reflect.Ptr:
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			elemValue := fieldValue.Elem()
			switch elemValue.Kind() {
			case reflect.String:
				elemValue.SetString(value)
			case reflect.Int:
				if i, err := strconv.Atoi(value); err == nil {
					elemValue.SetInt(int64(i))
				}
			case reflect.Bool:
				if b, err := strconv.ParseBool(value); err == nil {
					elemValue.SetBool(b)
				}
			}
		default:
			return fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}

	return nil
}
