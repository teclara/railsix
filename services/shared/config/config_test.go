package config

import (
	"os"
	"testing"
)

func TestEnvOr_WithEnvVar(t *testing.T) {
	const key = "TEST_ENVOR_SET"
	os.Setenv(key, "custom-value")
	defer os.Unsetenv(key)

	got := EnvOr(key, "fallback")
	if got != "custom-value" {
		t.Errorf("EnvOr(%q, %q) = %q, want %q", key, "fallback", got, "custom-value")
	}
}

func TestEnvOr_WithoutEnvVar(t *testing.T) {
	const key = "TEST_ENVOR_UNSET"
	os.Unsetenv(key)

	got := EnvOr(key, "fallback")
	if got != "fallback" {
		t.Errorf("EnvOr(%q, %q) = %q, want %q", key, "fallback", got, "fallback")
	}
}

func TestRequire_Success(t *testing.T) {
	const key = "TEST_REQUIRE_SET"
	os.Setenv(key, "required-value")
	defer os.Unsetenv(key)

	got, err := Require(key)
	if err != nil {
		t.Fatalf("Require(%q) returned unexpected error: %v", key, err)
	}
	if got != "required-value" {
		t.Errorf("Require(%q) = %q, want %q", key, got, "required-value")
	}
}

func TestRequire_Failure(t *testing.T) {
	const key = "TEST_REQUIRE_UNSET"
	os.Unsetenv(key)

	_, err := Require(key)
	if err == nil {
		t.Fatalf("Require(%q) expected error, got nil", key)
	}
}
