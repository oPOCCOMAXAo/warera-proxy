package internal

import "net/http"

type Service interface {
	Register(*http.ServeMux) error
}
