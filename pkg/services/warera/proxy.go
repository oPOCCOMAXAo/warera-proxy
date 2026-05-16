package warera

import (
	"net/http"

	"github.com/opoccomaxao/warera-proxy/pkg/utils/jsonutils"
)

func ProxyError(message string) []byte {
	return jsonutils.MarshalNoError(&Response{
		Error: &Error{
			Message: "proxy: " + message,
			Code:    ErrorCodeProxy,
			Data: ErrorData{
				Code:       TextErrorCodeProxy,
				HTTPStatus: http.StatusInternalServerError,
				Path:       "",
			},
		},
	})
}
