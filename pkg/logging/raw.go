package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// RawHandler just prints the message field as-is.
type rawHandler struct {
	w io.Writer
}

func (h *rawHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *rawHandler) Handle(_ context.Context, r slog.Record) error {
	if h.w == nil {
		h.w = os.Stdout
	}
	_, err := fmt.Fprint(h.w, r.Message)
	if err == nil {
		// slog.Record doesn't add its own newline,
		// so only add one if you want line breaks for each log call.
		_, _ = fmt.Fprint(h.w)
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

func NewRawLoggerTo(w io.Writer) *slog.Logger {
	return slog.New(&rawHandler{w: w})
}
