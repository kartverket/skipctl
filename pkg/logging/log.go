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

var (
	logger    *slog.Logger
	rawLogger *slog.Logger
	leveler   *slog.LevelVar
	lock      sync.Mutex

	ctxStdoutKey stdoutCtxKey

	DefaultStdoutContext = ForceStdoutContext(context.Background())
)

type stdoutCtxKey struct{}

type splitHandler struct {
	stdout, stderr slog.Handler
}

func (h *splitHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Delegate; assume same levels configured on both.
	return h.stdout.Enabled(ctx, level) || h.stderr.Enabled(ctx, level)
}

func (h *splitHandler) Handle(ctx context.Context, r slog.Record) error {
	toStdout, _ := ctx.Value(ctxStdoutKey).(bool)

	if toStdout {
		return h.stdout.Handle(ctx, r)
	}
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

	splitHandler := &splitHandler{stdout: newHandler(os.Stdout), stderr: newHandler(os.Stderr)}
	// slog-context outputs key-values found in the context to the log output
	ctxHandler := slogcontext.NewHandler(splitHandler)

	lock.Lock()
	defer lock.Unlock()
	logger = slog.New(ctxHandler)
	return logger
}

func ForceStdoutContext(parent context.Context) context.Context {
	return context.WithValue(parent, ctxStdoutKey, true)
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
