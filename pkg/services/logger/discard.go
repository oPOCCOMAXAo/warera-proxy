package logger

import "log/slog"

//nolint:gochecknoglobals
var discard = slog.New(slog.DiscardHandler)

func NewDiscard() *slog.Logger {
	return discard
}
