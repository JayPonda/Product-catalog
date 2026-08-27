package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Primitive interface {
	int | float64 | string | bool
}

type LoggerMeta = map[string]any

type Logger interface {
	Debug(file, method, message string, meta LoggerMeta)
	Info(file, method, message string, meta LoggerMeta)
	Warn(file, method, message string, meta LoggerMeta)
	Error(file, method, message string, meta LoggerMeta, trace string)
}

type StructuredLogger struct{}

var (
	singleTonLogger *StructuredLogger
	loggerOnce      sync.Once
)

func NewStructuredLogger() *StructuredLogger {
	loggerOnce.Do(func() {
		singleTonLogger = &StructuredLogger{}
	})
	return singleTonLogger
}

func (l *StructuredLogger) Debug(file, method, message string, meta LoggerMeta) {
	l.logWithLevel("-", "DEBUG", file, method, message, meta, "")
}

func (l *StructuredLogger) Info(file, method, message string, meta LoggerMeta) {
	l.logWithLevel("-", "INFO", file, method, message, meta, "")
}

func (l *StructuredLogger) Warn(file, method, message string, meta LoggerMeta) {
	l.logWithLevel("-", "WARN", file, method, message, meta, "")
}

func (l *StructuredLogger) Error(file, method, message string, meta LoggerMeta, trace string) {
	l.logWithLevel("-", "ERROR", file, method, message, meta, trace)
}

func (l *StructuredLogger) LogWithCorrelation(correlationID, level, file, method, message string, meta LoggerMeta) {
	l.logWithLevel(correlationID, level, file, method, message, meta, "")
}

func (l *StructuredLogger) logWithLevel(correlationID, level, file, method, message string, meta LoggerMeta, trace string) {
	ts := time.Now().Format("2006-01-02T15:04:05.000Z")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] [%s] %s %s.%s %s", ts, correlationID, level, file, method, message))

	if len(meta) > 0 {
		pairs := make([]string, 0, len(meta))
		for k, v := range meta {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}
		sb.WriteString(fmt.Sprintf(" %s", strings.Join(pairs, " ")))
	}

	if trace != "" {
		sb.WriteString(fmt.Sprintf(" trace=%s", trace))
	}

	fmt.Println(sb.String())
}

const LoggerContextKey = "app_logger"

func getLogger(ctx fiber.Ctx) Logger {
	if ctx != nil {
		if logger, ok := ctx.Locals(LoggerContextKey).(Logger); ok {
			return logger
		}
	}
	return NewStructuredLogger()
}
