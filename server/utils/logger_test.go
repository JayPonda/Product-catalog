package utils

import (
	"testing"
)

func TestStructuredLogger_NoPanicNilCtx(t *testing.T) {
	l := NewStructuredLogger()
	l.Debug(nil, "test.go", "TestFunc", "debug message", nil)
	l.Info(nil, "test.go", "TestFunc", "info message", LoggerMeta{"k": "v"})
	l.Warn(nil, "test.go", "TestFunc", "warn message", nil)
	l.Error(nil, "test.go", "TestFunc", "error message", nil, "trace")
}

func TestStructuredLogger_WithMeta(t *testing.T) {
	l := NewStructuredLogger()
	l.Info(nil, "file.go", "Method", "with meta", LoggerMeta{
		"count": 5,
		"name":  "test",
		"ok":    true,
	})
}

func TestNewStructuredLogger_Singleton(t *testing.T) {
	l1 := NewStructuredLogger()
	l2 := NewStructuredLogger()
	if l1 != l2 {
		t.Error("expected same singleton instance")
	}
}

func TestCLIContext_Locals(t *testing.T) {
	ctx := NewCLIContext()
	ctx.Locals(CorrelationIDKey, "test-123")
	if got := GetCorrelationID(ctx); got != "test-123" {
		t.Errorf("expected test-123, got %s", got)
	}
}

func TestCLIContext_NilContext(t *testing.T) {
	if got := GetCorrelationID(nil); got != "-" {
		t.Errorf("expected -, got %s", got)
	}
}
