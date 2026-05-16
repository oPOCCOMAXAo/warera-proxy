package lifecycle

import (
	"context"
	"time"

	"github.com/opoccomaxao/warera-proxy/pkg/services/logger"
)

type Package struct {
	Service *Service
}

type Config struct {
	// Shutdown timeout for each shutdownable service.
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

func MakePackage(
	cfg Config,
	appCancel context.CancelCauseFunc,
	logger *logger.Package,
) *Package {
	return &Package{
		Service: NewService(
			cfg,
			appCancel,
			logger.Logger,
		),
	}
}
