package config

import "testing"

func TestLoadUsesDefaultSessionSecretWhenUnset(t *testing.T) {
	t.Setenv("SESSION_SECRET", "")

	cfg := Load()
	if cfg.SessionSecret == "" {
		t.Fatal("expected default session secret when SESSION_SECRET is unset")
	}
}
