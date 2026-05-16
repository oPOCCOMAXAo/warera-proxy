package runtimeutils

import "reflect"

func GetFullTypeName(i any) string {
	typ := reflect.TypeOf(i)

	return typ.PkgPath() + " " + typ.String()
}
