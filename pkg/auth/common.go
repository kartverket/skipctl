package auth

import "log/slog"

var log *slog.Logger

func SetLogger(l *slog.Logger) {
	log = l
}
