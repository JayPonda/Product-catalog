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

func TestNewValidator_PrintableEmpty(t *testing.T) {
	v := NewValidator()
	// Empty string passes printable (no non-printable chars)
	if err := v.Var("", "printable"); err != nil {
		t.Errorf("empty string should pass printable: %v", err)
	}
}

func TestNewValidator_LetterUnicode(t *testing.T) {
	v := NewValidator()
	// Unicode letters should pass
	if err := v.Var("日本語", "letter"); err != nil {
		t.Errorf("unicode letters should pass: %v", err)
	}
	// Symbols only should fail
	if err := v.Var("!@#", "letter"); err == nil {
		t.Error("symbols only should fail letter check")
	}
}

func TestNewValidator_PrintableUnicode(t *testing.T) {
	v := NewValidator()
	// Unicode printable should pass
	if err := v.Var("日本語", "printable"); err != nil {
		t.Errorf("unicode printable should pass: %v", err)
	}
	// Space is printable
	if err := v.Var(" ", "printable"); err != nil {
		t.Errorf("space should be printable: %v", err)
	}
	// Tab is NOT printable (unicode.IsPrint returns false for \t)
	if err := v.Var("\t", "printable"); err == nil {
		t.Error("tab should NOT be printable")
	}
}
