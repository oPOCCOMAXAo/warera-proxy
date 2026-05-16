package logger

import "log/slog"

type Config struct {
	// Enable debug logging.
	Debug bool `env:"DEBUG" envDefault:"false"`
}

type Package struct {
	Logger *slog.Logger
}

func MakePackage(
	cfg Config,
) *Package {
	var res Package

	res.Logger = NewLogger(cfg)

	return &res
}

func (p *Package) Error(err error) {
	p.Logger.Error("Error",
		slog.Any("error", err),
	)
}
