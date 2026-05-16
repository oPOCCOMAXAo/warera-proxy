package api

import (
	"github.com/opoccomaxao/warera-proxy/pkg/api/internal"
	"github.com/opoccomaxao/warera-proxy/pkg/api/system"
	"github.com/opoccomaxao/warera-proxy/pkg/services/server"
)

func MakePackage(
	server *server.Package,
) error {
	svcs := []internal.Service{
		system.New(),
	}

	for _, svc := range svcs {
		err := svc.Register(server.Mux)
		if err != nil {
			return err
		}
	}

	return nil
}
