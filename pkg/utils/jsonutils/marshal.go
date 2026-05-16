package jsonutils

import "encoding/json"

//nolint:errchkjson
func MarshalNoError(object any) []byte {
	res, _ := json.Marshal(object)

	return res
}
