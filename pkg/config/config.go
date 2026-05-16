package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/opoccomaxao/warera-proxy/pkg/services/lifecycle"
	"github.com/opoccomaxao/warera-proxy/pkg/services/logger"
	"github.com/opoccomaxao/warera-proxy/pkg/services/server"
	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
	"github.com/pkg/errors"
)

type Config struct {
	Logger    logger.Config    `envPrefix:"LOGGER_"`
	Lifecycle lifecycle.Config `envPrefix:"LIFECYCLE_"`
	Server    server.Config    `envPrefix:"SERVER_"`
	Warera    warera.Config    `envPrefix:"WARERA_"`
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
