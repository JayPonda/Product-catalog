package utils

import (
	"sync"
	"testing"
)

func TestGetDB_NilContext(t *testing.T) {
	// Reset singleton for clean test
	dbOnce = sync.Once{}
	goquDBInstance = nil

	db := GetDB(nil)
	// Should return nil since not initialized
	if db != nil {
		t.Error("expected nil db when not initialized")
	}
}
