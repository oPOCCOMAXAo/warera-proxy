package lifecycle

import (
	"context"

	"github.com/opoccomaxao/warera-proxy/pkg/utils/runtimeutils"
)

type Servable interface {
	// Serve starts the servable component.
	//
	// It can block until the component is stopped or an error occurs.
	//
	// Context is app lifecycle context and will be cancelled when app shutdown is initiated.
	//
	// Returned error will be treated as a fatal error and will trigger shutdown.
	//
	// If Serve returns nil, app will continue to run until shutdown signal is received.
	Serve(context.Context) error
}

type Shutdownable interface {
	// Shutdown stops the shutdownable component.
	//
	// It must block until the component is fully stopped or an error occurs.
	//
	// Context is timeout context for the shutdown operation.
	//
	// Returned error will be logged.
	Shutdown(context.Context) error
}

type servableWithStack struct {
	Servable

	Stack *runtimeutils.Stack
}

type shutdownableWithStack struct {
	Shutdownable

	Stack *runtimeutils.Stack
}
