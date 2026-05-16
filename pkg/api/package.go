package api

import (
	"github.com/opoccomaxao/warera-proxy/pkg/api/internal"
	"github.com/opoccomaxao/warera-proxy/pkg/api/system"
	"github.com/opoccomaxao/warera-proxy/pkg/api/trpc"
	"github.com/opoccomaxao/warera-proxy/pkg/services/server"
	"github.com/opoccomaxao/warera-proxy/pkg/services/warera"
)

func MakePackage(
	server *server.Package,
	warera *warera.Package,
) error {
	svcs := []internal.Service{
		system.New(),
		trpc.NewService(
			warera.Client,
		),
	}

	for _, svc := range svcs {
		err := svc.Register(server.Mux)
		if err != nil {
			return err
		}
	}

	return nil
}
