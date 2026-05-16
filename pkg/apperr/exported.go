package apperr

import (
	"errors"
	"log/slog"
	"strings"
)

type ExportedError struct {
	OriginalError error
	Metadata      []slog.Attr
}

func NewExportedError(
	err error,
	attrs ...slog.Attr,
) error {
	if err == nil {
		panic("err must not be nil")
	}

	return &ExportedError{
		OriginalError: err,
		Metadata:      attrs,
	}
}

func (e *ExportedError) Error() string {
	var res strings.Builder

	res.WriteString(e.OriginalError.Error())

	for i, attr := range e.Metadata {
		if i == 0 {
			res.WriteString(" | metadata: ")
		} else {
			res.WriteString(", ")
		}

		res.WriteString(attr.Key)
		res.WriteString("=")
		res.WriteString(attr.Value.String())
	}

	return res.String()
}

func (e *ExportedError) Unwrap() error {
	return e.OriginalError
}

func AsExportedError(err error) (*ExportedError, bool) {
	return errors.AsType[*ExportedError](err)
}
