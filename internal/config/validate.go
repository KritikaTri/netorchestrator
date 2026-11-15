package config

import (
	"fmt"
	"os"
	"strings"
)

// AllowInsecureEnv, when set to "1", lets the server start with the built-in
// development secrets. It exists for local demos only and is logged loudly.
const AllowInsecureEnv = "NETORCH_ALLOW_INSECURE_DEFAULTS"

const minSecretLen = 32

// knownDefaultSecrets are the placeholder values that used to ship as defaults.
var knownDefaultSecrets = []string{
	"your-secret-key",
	"netorchestrator-super-secret-jwt-key-2024",
	"netorchestrator-encryption-key-32-byte",
	"changeme",
	"secret",
}

// Validate refuses to run with a guessable JWT secret. It returns a
// non-empty warning (and no error) when the insecure escape hatch is used.
func (c *Config) Validate() (warning string, err error) {
	problem := secretProblem(c.Security.JWTSecret)
	if problem == "" {
		return "", nil
	}
	if os.Getenv(AllowInsecureEnv) == "1" {
		return fmt.Sprintf("INSECURE: security.jwt_secret %s; allowed only because %s=1", problem, AllowInsecureEnv), nil
	}
	return "", fmt.Errorf("security.jwt_secret %s; set SECURITY_JWT_SECRET to a random value of at least %d characters (or %s=1 for local demos only)", problem, minSecretLen, AllowInsecureEnv)
}

func secretProblem(s string) string {
	if strings.TrimSpace(s) == "" {
		return "is empty"
	}
	for _, d := range knownDefaultSecrets {
		if strings.EqualFold(s, d) {
			return "is a well-known default"
		}
	}
	if len(s) < minSecretLen {
		return fmt.Sprintf("is shorter than %d characters", minSecretLen)
	}
	return ""
}
