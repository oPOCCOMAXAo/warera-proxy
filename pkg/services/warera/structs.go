package warera

type Response struct {
	Error *Error `json:"error,omitempty"`
}

type Error struct {
	Message string    `json:"message"`
	Code    int       `json:"code"`
	Data    ErrorData `json:"data"`
}

type ErrorData struct {
	Code       string `json:"code"`
	HTTPStatus int    `json:"httpStatus"`
	Path       string `json:"path"`
}
