package utils

import (
	"testing"
)

func TestNewValidator_Letter(t *testing.T) {
	v := NewValidator()
	if err := v.Var("abc123", "letter"); err != nil {
		t.Errorf("expected 'abc123' to pass letter check: %v", err)
	}
	if err := v.Var("123", "letter"); err == nil {
		t.Error("expected digits-only to fail letter check")
	}
}

func TestNewValidator_Printable(t *testing.T) {
	v := NewValidator()
	if err := v.Var("hello", "printable"); err != nil {
		t.Errorf("unexpected error for printable: %v", err)
	}
	if err := v.Var("hi\x00", "printable"); err == nil {
		t.Error("expected control char to fail printable check")
	}
}

func TestNewValidator_Struct(t *testing.T) {
	v := NewValidator()
	type sample struct {
		Name string `validate:"letter"`
	}
	s := sample{Name: "123"}
	if err := v.Struct(s); err == nil {
		t.Error("expected validation error for Name without letters")
	}
}
