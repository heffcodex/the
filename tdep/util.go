package tdep

import (
	"reflect"
)

func typeOfT[T any]() string {
	tof := reflect.TypeFor[T]()
	if tof.Kind() == reflect.Pointer {
		tof = tof.Elem()
	}

	return tof.String()
}
