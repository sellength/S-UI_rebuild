package main

import "testing"

func TestParseSingboxVersion(t *testing.T) {
	cases := map[string]string{
		"sing-box version 1.12.0\nEnvironment: go1.24": "1.12.0",
		"v1.13.1": "1.13.1",
		"":        "unknown",
	}

	for input, expected := range cases {
		if got := parseSingboxVersion(input); got != expected {
			t.Fatalf("parseSingboxVersion(%q) = %q, want %q", input, got, expected)
		}
	}
}
