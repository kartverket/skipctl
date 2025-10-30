package telemetry

import (
	"fmt"
	"log/slog"
)

type posthogSlogAdapter struct {
	l *slog.Logger
}

func (s *posthogSlogAdapter) Debugf(format string, args ...interface{}) {
	s.l.Debug(fmt.Sprintf(format, args...))
}

func (s *posthogSlogAdapter) Logf(format string, args ...interface{}) {
	s.l.Info(fmt.Sprintf(format, args...))
}

func (s *posthogSlogAdapter) Warnf(format string, args ...interface{}) {
	s.l.Warn(fmt.Sprintf(format, args...))
}

func (s *posthogSlogAdapter) Errorf(format string, args ...interface{}) {
	s.l.Error(fmt.Sprintf(format, args...))
}
