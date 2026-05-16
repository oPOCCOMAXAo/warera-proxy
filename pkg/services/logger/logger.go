package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
)

func NewLogger(
	cfg Config,
) *slog.Logger {
	if cfg.Debug {
		return NewDebugLogger()
	}

	return NewProductionLogger()
}

//nolint:mnd
func NewDebugLogger() *slog.Logger {
	return slog.New(devslog.NewHandler(os.Stdout, &devslog.Options{
		HandlerOptions: &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: replacer,
		},
		MaxErrorStackTrace: 5,
	}))
}

func NewProductionLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: replacer,
	}))
}

//nolint:forbidigo
func replacer(groups []string, attr slog.Attr) slog.Attr {
	if attr.Key == "error" {
		err, ok := attr.Value.Any().(error)
		if ok {
			return slog.String(
				attr.Key,
				fmt.Sprintf("%+v", err),
			)
		}
	}

	return attr
}
