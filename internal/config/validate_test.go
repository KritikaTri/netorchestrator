package config

import (
	"strings"
	"testing"
)

func cfgWith(secret string) *Config {
	c := &Config{}
	c.Security.JWTSecret = secret
	return c
}

func TestValidateRejectsWeakSecrets(t *testing.T) {
	t.Setenv(AllowInsecureEnv, "")
	for _, s := range []string{"", "   ", "your-secret-key", "netorchestrator-super-secret-jwt-key-2024", "short", "CHANGEME"} {
		if _, err := cfgWith(s).Validate(); err == nil {
			t.Errorf("secret %q should be rejected", s)
		}
	}
}

func TestValidateAcceptsStrongSecret(t *testing.T) {
	t.Setenv(AllowInsecureEnv, "")
	w, err := cfgWith(strings.Repeat("x7Qp", 10)).Validate()
	if err != nil || w != "" {
		t.Fatalf("w=%q err=%v", w, err)
	}
}

func TestEscapeHatchWarnsButAllows(t *testing.T) {
	t.Setenv(AllowInsecureEnv, "1")
	w, err := cfgWith("your-secret-key").Validate()
	if err != nil || !strings.Contains(w, "INSECURE") {
		t.Fatalf("w=%q err=%v", w, err)
	}
}
