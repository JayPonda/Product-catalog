package utils

import (
	"testing"
)

func TestStructuredLogger_NoPanicNilCtx(t *testing.T) {
	l := NewStructuredLogger()
	l.Debug("test.go", "TestFunc", "debug message", nil)
	l.Info("test.go", "TestFunc", "info message", LoggerMeta{"k": "v"})
	l.Warn("test.go", "TestFunc", "warn message", nil)
	l.Error("test.go", "TestFunc", "error message", nil, "trace")
}

func TestStructuredLogger_WithMeta(t *testing.T) {
	l := NewStructuredLogger()
	l.Info("file.go", "Method", "with meta", LoggerMeta{
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
