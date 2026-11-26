package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	slogcontext "github.com/PumpkinSeed/slog-context"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/yannh/kubeconform/pkg/validator"
)

var (
	logger     *slog.Logger
	rawLogger  *slog.Logger
	leveler    *slog.LevelVar
	lock       sync.Mutex
	outputMode OutputMode

	ctxStdoutKey stdoutCtxKey

	DefaultStdoutContext = ForceStdoutContext(context.Background())
)

type stdoutCtxKey struct{}

type splitHandler struct {
	stdout, stderr slog.Handler
}

// Style used for error messages. Red background with white text.
var errStyle = color.New(color.FgWhite, color.BgRed).SprintFunc()

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

	outputMode = parsedMode

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

func logValidationErrorsJSON(filename string, validationErrors []validator.ValidationError, err error) {
	// JSON mode: only log structured errors
	if err != nil {
		logger.Error("validation error", "file", filename, "error", err.Error())
	}
	if len(validationErrors) > 0 {
		errorMsgs := make([]string, len(validationErrors))
		for i, ve := range validationErrors {
			errorMsgs[i] = fmt.Sprintf("{%s %s}", ve.Path, strings.Trim(ve.Error(), "{}"))
		}
		logger.Error("validation errors", "file", filename, "errors", errorMsgs)
	}
}

func logValidationErrorsText(filename string, validationErrors []validator.ValidationError, err error) {
	// Text mode: human-readable formatting
	rawLogger.Error(
		errStyle("ERROR:") + fmt.Sprintf(" file is invalid at %s", filename),
	)

	// Print each validation error
	for _, ve := range validationErrors {
		cleanedMsg := strings.Trim(ve.Error(), "{}") // remove outer braces
		rawLogger.Error(fmt.Sprintf("  — %s: %s\n", ve.Path, cleanedMsg))
	}

	if err != nil {
		errMsg := strings.TrimSpace(err.Error())
		// Split error message to show hint on separate lines for better readability
		if strings.Contains(errMsg, "Hint:") {
			var maxHintParts = 2
			parts := strings.SplitN(errMsg, "Hint:", maxHintParts)
			rawLogger.Error(fmt.Sprintf("  — %s\n", parts[0]))
			rawLogger.Error(fmt.Sprintf("   Hint:%s\n", parts[1]))
		} else {
			rawLogger.Error(fmt.Sprintf("  — %s\n", errMsg))
		}
	}
}

func LogValidationErrors(filename string, validationErrors []validator.ValidationError, err error, outputJSON bool) {
	// Skip colored raw output when in JSON mode - errors are already logged via the structured logger
	if outputMode == OutputModeJSON {
		return
	}

	if rawLogger == nil {
		panic("logger not initialized")
	}

	if validationErrors == nil && err == nil {
		return
	}

	if outputJSON {
		logValidationErrorsJSON(filename, validationErrors, err)
	} else {
		logValidationErrorsText(filename, validationErrors, err)
	}
}
