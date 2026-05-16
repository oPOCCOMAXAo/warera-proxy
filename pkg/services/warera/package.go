package warera

import (
	"log/slog"

	"github.com/opoccomaxao/warera-proxy/pkg/services/lifecycle"
	"github.com/opoccomaxao/warera-proxy/pkg/services/logger"
)

type Config struct {
	Token             string  `env:"TOKEN,required"`
	Host              string  `env:"HOST"                envDefault:"https://api2.warera.io"`
	BatchSize         int     `env:"BATCH_SIZE"          envDefault:"50"`
	RequestsPerMinute float64 `env:"REQUESTS_PER_MINUTE" envDefault:"200.0"`
}

type Package struct {
	Client *Client
}

func MakePackage(
	config Config,
	lifecycle *lifecycle.Package,
	logger *logger.Package,
) *Package {
	var res Package

	res.Client = NewClient(
		config,
		logger.Logger.With(
			slog.String("service", "warera"),
		),
	)

	lifecycle.Service.RegisterService(res.Client)

	return &res
}
