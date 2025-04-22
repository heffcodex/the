package tdep

import (
	"reflect"
)

func typeOfT[T any]() string {
	tof := reflect.TypeOf(new(T)).Elem()
	if tof.Kind() == reflect.Pointer {
		tof = tof.Elem()
	}

	return tof.String()
}
