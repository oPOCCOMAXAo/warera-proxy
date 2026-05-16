package lifecycle

import (
	"context"
	"log/slog"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/opoccomaxao/warera-proxy/pkg/utils/runtimeutils"
)

type Service struct {
	mu sync.Mutex

	appCancel       context.CancelCauseFunc
	shutdownTimeout time.Duration
	shutdownable    []*shutdownableWithStack
	servable        []*servableWithStack

	logger *slog.Logger
}

func NewService(
	cfg Config,
	appCancel context.CancelCauseFunc,
	logger *slog.Logger,
) *Service {
	return &Service{
		appCancel:       appCancel,
		shutdownTimeout: cfg.ShutdownTimeout,
		logger:          logger,
	}
}

// RegisterService registers a service for lifecycle management.
func (s *Service) RegisterService(service any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullName := runtimeutils.GetFullTypeName(service)
	stack := runtimeutils.CaptureStack(1)

	hasAny := false

	if shutdownable, ok := service.(Shutdownable); ok {
		s.shutdownable = append(s.shutdownable, &shutdownableWithStack{
			Shutdownable: shutdownable,
			Stack:        stack,
		})

		s.logger.Info("Registered Shutdownable Service",
			slog.String("service", fullName),
			stack.SlogLines(),
		)

		hasAny = true
	}

	if servable, ok := service.(Servable); ok {
		s.servable = append(s.servable, &servableWithStack{
			Servable: servable,
			Stack:    stack,
		})

		s.logger.Info("Registered Servable Service",
			slog.String("service", fullName),
			stack.SlogLines(),
		)

		hasAny = true
	}

	if !hasAny {
		s.logger.Warn("Service does not implement any lifecycle interface",
			slog.String("service", fullName),
			stack.SlogLines(),
		)
	}
}

// Serve starts all registered servable components.
//
// Non-blocking.
func (s *Service) Serve(
	ctx context.Context,
	cancelCause context.CancelCauseFunc,
) error {
	s.logger.InfoContext(ctx, "lifecycle.Serve",
		slog.String("go_version", runtime.Version()),
	)

	for _, servable := range s.servable {
		go func() {
			err := servable.Serve(ctx)
			if err != nil {
				s.logger.Error("Serve Failed",
					slog.Any("error", err),
					servable.Stack.SlogLines(),
				)

				cancelCause(err)
			}
		}()
	}

	return nil
}

// Shutdown stops all registered shutdownable components.
//
// Blocks until all components are stopped or an error occurs.
//
//nolint:contextcheck
func (s *Service) Shutdown(_ context.Context) error {
	for _, shutdownable := range slices.Backward(s.shutdownable) {
		func() {
			actionCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
			defer cancel()

			err := shutdownable.Shutdown(actionCtx)
			if err != nil {
				s.logger.Error("Shutdown Failed",
					slog.Any("error", err),
					shutdownable.Stack.SlogLines(),
				)
			}
		}()
	}

	return nil
}

func (s *Service) shutdownApp(
	_ context.Context,
	err error,
	stack *runtimeutils.Stack,
) {
	if stack == nil {
		stack = runtimeutils.CaptureStack(1)
	}

	var attrs []any

	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	attrs = append(attrs, stack.SlogLines())

	s.logger.Info("Application Shutdown Requested", attrs...)

	s.appCancel(err)
}

func (s *Service) ShutdownAppWithError(
	ctx context.Context,
	err error,
) {
	stack := runtimeutils.CaptureStack(1)

	s.shutdownApp(ctx, err, stack)
}

func (s *Service) ShutdownApp(ctx context.Context) {
	stack := runtimeutils.CaptureStack(1)

	s.shutdownApp(ctx, nil, stack)
}
