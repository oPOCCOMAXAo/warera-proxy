package server

import (
	"log/slog"
	"net/http"

	"github.com/opoccomaxao/warera-proxy/pkg/services/lifecycle"
	"github.com/opoccomaxao/warera-proxy/pkg/services/logger"
)

type Config struct {
	Port int64 `default:"8080" env:"PORT"`
}

type Package struct {
	Service *Service
	Mux     *http.ServeMux
}

func MakePackage(
	config Config,
	lifecycle *lifecycle.Package,
	logger *logger.Package,
) *Package {
	var res Package

	res.Service = NewService(
		config,
		logger.Logger.With(
			slog.String("service", "server"),
		),
	)

	res.Mux = res.Service.Mux()

	lifecycle.Service.RegisterService(res.Service)

	return &res
}
