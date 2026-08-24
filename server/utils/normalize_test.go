package utils

import "testing"

func TestNormalizeName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"  Hello World ", "hello world"},
		{"FOO", "foo"},
		{"  bar  ", "bar"},
		{"", ""},
		{"\t trimmed \n", "trimmed"},
	}
	for _, c := range cases {
		if got := NormalizeName(c.in); got != c.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
