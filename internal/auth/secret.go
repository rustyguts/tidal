package auth

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HeaderName is the header used to authenticate /api/internal/* callbacks.
const HeaderName = "X-Tidal-Callback-Secret"

// CallbackSecret returns Echo middleware that requires a constant-time match
// of the configured shared secret. If secret is empty, it 503s — refuses to
// run open. Server bootstrap should fail before reaching here when k8s mode
// is selected without a secret.
func CallbackSecret(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if secret == "" {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "callback secret not configured")
			}
			got := c.Request().Header.Get(HeaderName)
			if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid callback secret")
			}
			return next(c)
		}
	}
}
