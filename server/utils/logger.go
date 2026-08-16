package utils

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/gofiber/fiber/v3"
)

// 1. Core type definitions
type Primitive interface {
	int | float64 | string | bool
}

type LoggerMeta = map[string]any

type Logger interface {
	Debug(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta)
	Info(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta)
	Warn(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta)
	Error(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta, trace string)
}

// 2. Struct implementing the Logger interface
type StructuredLogger struct {
	logger *slog.Logger
}

var (
	singleTonLogger *StructuredLogger
	loggerOnce      sync.Once // Ensures thread-safe, one-time execution
)

// NewStructuredLogger initializes a structured JSON logger safely using sync.Once
func NewStructuredLogger() *StructuredLogger {
	loggerOnce.Do(func() {
		singleTonLogger = &StructuredLogger{
			logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})),
		}
	})

	return singleTonLogger
}

// 3. Interface Method Implementations

func (l *StructuredLogger) Debug(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta) {
	l.logWithLevel(ctx, slog.LevelDebug, source, identifier, message, meta, "")
}

func (l *StructuredLogger) Info(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta) {
	l.logWithLevel(ctx, slog.LevelInfo, source, identifier, message, meta, "")
}

func (l *StructuredLogger) Warn(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta) {
	l.logWithLevel(ctx, slog.LevelWarn, source, identifier, message, meta, "")
}

func (l *StructuredLogger) Error(ctx fiber.Ctx, source, identifier, message string, meta LoggerMeta, trace string) {
	l.logWithLevel(ctx, slog.LevelError, source, identifier, message, meta, trace)
}

// Helper method to eliminate duplicate boilerplate code across log calls
func (l *StructuredLogger) logWithLevel(ctx fiber.Ctx, level slog.Level, source, identifier, message string, meta LoggerMeta, trace string) {
	var goCtx context.Context
	if ctx != nil {
		goCtx = ctx
	} else {
		goCtx = context.Background()
	}

	args := []any{
		slog.String("source", source),
		slog.String("identifier", identifier),
	}

	if len(meta) > 0 {
		metaArgs := make([]any, 0, len(meta)*2)
		for k, v := range meta {
			metaArgs = append(metaArgs, k, v)
		}
		args = append(args, slog.Group("meta", metaArgs...))
	}

	if trace != "" {
		args = append(args, slog.String("trace", trace))
	}

	l.logger.Log(goCtx, level, message, args...)
}

// 4. Thread-Safe Dependency Injection Setup
const LoggerContextKey = "app_logger"

func getLogger(ctx fiber.Ctx) Logger {
	if ctx != nil {
		if logger, ok := ctx.Locals(LoggerContextKey).(Logger); ok {
			return logger
		}
	}
	return NewStructuredLogger()
}
