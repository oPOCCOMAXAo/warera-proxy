package app

import (
	"context"
	"os"
	"os/signal"

	"github.com/opoccomaxao/warera-proxy/pkg/api"
	"github.com/opoccomaxao/warera-proxy/pkg/config"
	"github.com/opoccomaxao/warera-proxy/pkg/services/lifecycle"
	"github.com/opoccomaxao/warera-proxy/pkg/services/logger"
	"github.com/opoccomaxao/warera-proxy/pkg/services/server"
	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
)

func Run() error {
	appCtx, appCancelCause := context.WithCancelCause(context.Background())
	defer appCancelCause(nil)

	appCtx, cancel := signal.NotifyContext(appCtx, os.Interrupt)
	defer cancel()

	config, err := config.Load()
	if err != nil {
		return err
	}

	logger := logger.MakePackage(config.Logger)

	lifecycle := lifecycle.MakePackage(
		config.Lifecycle,
		appCancelCause,
		logger,
	)

	server := server.MakePackage(
		config.Server,
		lifecycle,
		logger,
	)

	warera := warera.MakePackage(
		config.Warera,
		lifecycle,
		logger,
	)

	err = api.MakePackage(
		server,
		warera,
	)
	if err != nil {
		return err
	}

	err = lifecycle.Service.Serve(appCtx, appCancelCause)
	if err != nil {
		logger.Error(err)

		return err
	}

	<-appCtx.Done()

	err = lifecycle.Service.Shutdown(appCtx)
	if err != nil {
		logger.Error(err)

		return err
	}

	return nil
}
