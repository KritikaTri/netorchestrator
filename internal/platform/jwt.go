package platform

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Claims is the subset of JWT claims the platform uses.
type Claims struct {
	Subject string   `json:"sub"`
	Roles   []string `json:"roles,omitempty"`
	Expires int64    `json:"exp"`
}

// ErrBadToken is returned for any invalid token; the reason is intentionally
// not exposed to callers.
var ErrBadToken = errors.New("platform: invalid token")

var b64 = base64.RawURLEncoding

// SignHS256 issues a compact HS256 JWT. Used by tests and the dev token tool.
func SignHS256(secret []byte, c Claims) string {
	h := b64.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	p, _ := json.Marshal(c)
	signing := h + "." + b64.EncodeToString(p)
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(signing))
	return signing + "." + b64.EncodeToString(m.Sum(nil))
}

// VerifyHS256 validates signature, algorithm and expiry. Only HS256 is
// accepted ("alg":"none" and algorithm-confusion tokens are rejected).
func VerifyHS256(secret []byte, token string, now time.Time) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrBadToken
	}
	hdrRaw, err := b64.DecodeString(parts[0])
	if err != nil {
		return nil, ErrBadToken
	}
	var hdr struct {
		Alg string `json:"alg"`
	}
	if json.Unmarshal(hdrRaw, &hdr) != nil || hdr.Alg != "HS256" {
		return nil, ErrBadToken
	}
	sig, err := b64.DecodeString(parts[2])
	if err != nil {
		return nil, ErrBadToken
	}
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, m.Sum(nil)) {
		return nil, ErrBadToken
	}
	body, err := b64.DecodeString(parts[1])
	if err != nil {
		return nil, ErrBadToken
	}
	var c Claims
	if json.Unmarshal(body, &c) != nil || c.Subject == "" || c.Expires == 0 || now.Unix() >= c.Expires {
		return nil, ErrBadToken
	}
	return &c, nil
}
