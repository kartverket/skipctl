package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"

	slogcontext "github.com/PumpkinSeed/slog-context"
	"github.com/pkg/errors"
)

var logger *slog.Logger
var rawLogger *slog.Logger
var leveler *slog.LevelVar

var lock sync.Mutex

type splitHandler struct {
	stdout slog.Handler
	stderr slog.Handler
}

func (h *splitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// All routing happens in Handle; Enabled just defers to one handler (options identical).
	return h.stdout.Enabled(ctx, level)
}

func (h *splitHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level == slog.LevelInfo {
		return h.stdout.Handle(ctx, r)
	}
	// Warn, Error (and optionally Debug if enabled) go to stderr.
	return h.stderr.Handle(ctx, r)
}

func (h *splitHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &splitHandler{
		stdout: h.stdout.WithAttrs(attrs),
		stderr: h.stderr.WithAttrs(attrs),
	}
}

func (h *splitHandler) WithGroup(name string) slog.Handler {
	return &splitHandler{
		stdout: h.stdout.WithGroup(name),
		stderr: h.stderr.WithGroup(name),
	}
}

func init() {
	rawLogger = slog.New(&rawHandler{})
	leveler = new(slog.LevelVar)
	leveler.Set(slog.LevelInfo)
	// This is the default before anyone calls this function
	ConfigureLogging("json", false)
}
func ConfigureLogging(mode string, isDebug bool) *slog.Logger {
	parsedMode, err := parseOutputMode(mode)
	if err != nil {
		panic(err)
	}

	if isDebug {
		leveler.Set(slog.LevelDebug)
	}

	opts := &slog.HandlerOptions{
		Level:     leveler,
		AddSource: isDebug,
	}

	newHandler := func(w io.Writer) slog.Handler {
		switch parsedMode {
		case OutputModeJSON:
			return slog.NewJSONHandler(w, opts)
		case OutputModeText:
			return slog.NewTextHandler(w, opts)
		default:
			panic(errors.Errorf("invalid output option: %v", parsedMode))
		}
	}
	stdoutBase := newHandler(os.Stdout)
	stderrBase := newHandler(os.Stderr)
	splitHandler := &splitHandler{stdout: stdoutBase, stderr: stderrBase}
	// slog-context outputs key-values found in the context to the log output
	ctxHandler := slogcontext.NewHandler(splitHandler)

	lock.Lock()
	defer lock.Unlock()
	logger = slog.New(ctxHandler)
	return logger
}

func Logger() *slog.Logger {
	if logger == nil {
		panic("logger not initialized")
	}

	return logger
}

// RawLogger returns a logger that just prints the message field as-is, without any extra formatting.
// This is useful for programs that need to output raw data.
func RawLogger() *slog.Logger {
	if rawLogger == nil {
		panic("logger not initialized")
	}

	return rawLogger
}
