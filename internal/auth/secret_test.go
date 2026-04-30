package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCallbackSecret_Valid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderName, "my-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CallbackSecret("my-secret")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCallbackSecret_Invalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderName, "wrong-secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CallbackSecret("my-secret")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("code = %d, want %d", he.Code, http.StatusUnauthorized)
	}
}

func TestCallbackSecret_EmptySecret(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CallbackSecret("")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if he.Code != http.StatusServiceUnavailable {
		t.Errorf("code = %d, want %d", he.Code, http.StatusServiceUnavailable)
	}
}

func TestCallbackSecret_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CallbackSecret("my-secret")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		t.Fatal("expected error")
	}
}
