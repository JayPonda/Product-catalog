package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestGetUUID(t *testing.T) {
	id, err := GetUUID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == uuid.Nil {
		t.Error("expected non-nil uuid")
	}
	if id.Version() != 1 {
		t.Errorf("expected v1 uuid, got v%d", id.Version())
	}
}
