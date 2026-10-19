package util

import "testing"

func TestGetEnv(t *testing.T) {
	t.Setenv("BUZZ_TEST_VALUE", "set")
	if got := GetEnv("BUZZ_TEST_VALUE", "fallback"); got != "set" {
		t.Fatalf("GetEnv() = %q, want set", got)
	}
	t.Setenv("BUZZ_TEST_EMPTY", "")
	if got := GetEnv("BUZZ_TEST_EMPTY", "fallback"); got != "" {
		t.Fatalf("GetEnv() = %q, want empty value", got)
	}
	if got := GetEnv("BUZZ_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("GetEnv() = %q, want fallback", got)
	}
}
