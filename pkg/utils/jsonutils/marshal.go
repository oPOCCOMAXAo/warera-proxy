package jsonutils

import "encoding/json"

func MarshalNoError(object any) []byte {
	res, _ := json.Marshal(object)

	return res
}
