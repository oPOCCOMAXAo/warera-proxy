package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/opoccomaxao/warera-proxy/pkg/services/lifecycle"
	"github.com/opoccomaxao/warera-proxy/pkg/services/logger"
	"github.com/pkg/errors"
)

type Config struct {
	Logger    logger.Config    `envPrefix:"LOGGER_"`
	Lifecycle lifecycle.Config `envPrefix:"LIFECYCLE_"`
}

func Load() (*Config, error) {
	var res Config

	err := env.ParseWithOptions(&res, env.Options{
		RequiredIfNoDef:       false,
		UseFieldNameByDefault: false,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &res, nil
}
