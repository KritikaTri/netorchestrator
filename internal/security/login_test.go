package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func login(t *testing.T, user, pass string) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", NewAuthHandlers(newMgr()).Login)
	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"`+user+`","password":"`+pass+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

// Regression: admin/admin123 used to log in on any deployment.
func TestDemoCredentialsDisabledByDefault(t *testing.T) {
	t.Setenv("NETORCH_ENABLE_DEMO_AUTH", "")
	if c := login(t, "admin", "admin123"); c != http.StatusUnauthorized {
		t.Fatalf("demo admin login must be refused by default, got %d", c)
	}
}

func TestDemoCredentialsWorkOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Setenv("NETORCH_ENABLE_DEMO_AUTH", "1")
	if c := login(t, "admin", "admin123"); c != http.StatusOK {
		t.Errorf("enabled demo login: %d", c)
	}
	if c := login(t, "admin", "wrong"); c != http.StatusUnauthorized {
		t.Errorf("wrong password: %d", c)
	}
}
