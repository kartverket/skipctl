package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// RawHandler just prints the message field as-is.
type rawHandler struct{}

func (h *rawHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *rawHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := fmt.Fprint(os.Stdout, r.Message)
	if err == nil {
		// slog.Record doesn't add its own newline,
		// so only add one if you want line breaks for each log call.
		_, _ = fmt.Fprintln(os.Stdout)
	}
	return err
}

func (h *rawHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// ignore attributes in this minimal implementation
	return h
}

func (h *rawHandler) WithGroup(_ string) slog.Handler {
	// ignore groups in this minimal implementation
	return h
}
