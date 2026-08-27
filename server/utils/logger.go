package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Primitive interface {
	int | float64 | string | bool
}

type LoggerMeta = map[string]any

// RequestContext abstracts request-scoped data (correlation ID, user ID, etc.).
// fiber.Ctx already satisfies this via its Locals method.
type RequestContext interface {
	Locals(key any, value ...any) any
}

// CLIContext implements RequestContext for non-HTTP contexts (CLI commands).
type CLIContext struct {
	locals map[string]any
}

func NewCLIContext() *CLIContext {
	return &CLIContext{locals: make(map[string]any)}
}

func (c *CLIContext) Locals(key any, value ...any) any {
	if len(value) > 0 {
		c.locals[key.(string)] = value[0]
	}
	return c.locals[key.(string)]
}

type Logger interface {
	Debug(ctx RequestContext, file, method, message string, meta LoggerMeta)
	Info(ctx RequestContext, file, method, message string, meta LoggerMeta)
	Warn(ctx RequestContext, file, method, message string, meta LoggerMeta)
	Error(ctx RequestContext, file, method, message string, meta LoggerMeta, trace string)
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

func (l *StructuredLogger) Debug(ctx RequestContext, file, method, message string, meta LoggerMeta) {
	l.logWithLevel(getCorrelationID(ctx), "DEBUG", file, method, message, meta, "")
}

func (l *StructuredLogger) Info(ctx RequestContext, file, method, message string, meta LoggerMeta) {
	l.logWithLevel(getCorrelationID(ctx), "INFO", file, method, message, meta, "")
}

func (l *StructuredLogger) Warn(ctx RequestContext, file, method, message string, meta LoggerMeta) {
	l.logWithLevel(getCorrelationID(ctx), "WARN", file, method, message, meta, "")
}

func (l *StructuredLogger) Error(ctx RequestContext, file, method, message string, meta LoggerMeta, trace string) {
	l.logWithLevel(getCorrelationID(ctx), "ERROR", file, method, message, meta, trace)
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

const (
	CorrelationIDKey = "correlation_id"
	LoggerContextKey = "app_logger"
)

func getCorrelationID(ctx RequestContext) string {
	if ctx != nil {
		if id, ok := ctx.Locals(CorrelationIDKey).(string); ok {
			return id
		}
	}
	return "-"
}

func GetCorrelationID(ctx RequestContext) string {
	return getCorrelationID(ctx)
}
