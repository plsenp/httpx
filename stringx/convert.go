package stringx

import (
	"errors"
	"reflect"
	"strconv"
	"time"
)

func To[T Convertable](s string) (T, error) {
	var result T
	val := reflect.ValueOf(&result).Elem()
	switch val.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		res, err := strconv.ParseInt(s, 10, val.Type().Bits())
		if err != nil {
			return result, err
		}
		val.SetInt(res)
		return result, nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		res, err := strconv.ParseUint(s, 10, val.Type().Bits())
		if err != nil {
			return result, err
		}
		val.SetUint(res)
		return result, nil
	case reflect.Float64, reflect.Float32:
		res, err := strconv.ParseFloat(s, val.Type().Bits())
		if err != nil {
			return result, err
		}
		val.SetFloat(res)
		return result, nil
	case reflect.String:
		val.SetString(s)
		return result, nil
	case reflect.Struct:
		if _, ok := any(result).(time.Time); ok {
			t, err := time.Parse(time.DateTime, s)
			if err != nil {
				return result, err
			}
			val.Set(reflect.ValueOf(t))
			return result, nil
		}
	}
	return result, errors.New("unsupported type")
}

func ToSlice[T Convertable](s []string) ([]T, error) {
	var result []T
	for _, v := range s {
		res, err := To[T](v)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return result, nil
}

type Convertable interface {
	~int | ~int64 | ~int32 | ~int16 | ~int8 |
		~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8 |
		~float32 | ~float64 |
		~string |
		time.Time
}
