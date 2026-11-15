package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newMgr() *SecurityManager { return NewSecurityManager(strings.Repeat("k", 40)) }

func authedRouter(sm *SecurityManager) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/who", sm.APIKeyAuthMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("user_id"))
	})
	return r
}

func call(r *gin.Engine, key string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/who", nil)
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Regression: any string starting with "neto_" used to authenticate.
func TestForgedNetoPrefixKeyIsRejected(t *testing.T) {
	r := authedRouter(newMgr())
	for _, k := range []string{"neto_anything", "neto_", "neto_" + strings.Repeat("0", 64), ""} {
		if w := call(r, k); w.Code != http.StatusUnauthorized {
			t.Errorf("key %q: got %d, want 401", k, w.Code)
		}
	}
}

func TestIssuedKeyAuthenticatesAsItsOwner(t *testing.T) {
	sm := newMgr()
	k, err := sm.APIKeyManager().GenerateAPIKey("ci", "user-42", []string{PermissionReadNetworks}, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := call(authedRouter(sm), k.Key)
	if w.Code != http.StatusOK || w.Body.String() != "user-42" {
		t.Fatalf("got %d %q", w.Code, w.Body)
	}
}

func TestRevokedAndExpiredKeysAreRejected(t *testing.T) {
	sm := newMgr()
	m := sm.APIKeyManager()
	r := authedRouter(sm)

	k, _ := m.GenerateAPIKey("a", "u", nil, nil)
	if err := m.RevokeAPIKey(k.ID); err != nil {
		t.Fatal(err)
	}
	if w := call(r, k.Key); w.Code != http.StatusUnauthorized {
		t.Errorf("revoked key accepted: %d", w.Code)
	}

	past := time.Now().Add(-time.Minute)
	e, _ := m.GenerateAPIKey("b", "u", nil, &past)
	if w := call(r, e.Key); w.Code != http.StatusUnauthorized {
		t.Errorf("expired key accepted: %d", w.Code)
	}
	if err := m.RevokeAPIKey("nope"); err == nil {
		t.Error("revoking unknown id should error")
	}
}

func TestPlaintextKeyIsNotRetainedServerSide(t *testing.T) {
	m := newMgr().APIKeyManager()
	k, _ := m.GenerateAPIKey("a", "u", nil, nil)
	for _, rec := range m.byID {
		if rec.Key != "" || strings.Contains(rec.HashedKey, k.Key) {
			t.Fatal("plaintext key retained")
		}
	}
}
