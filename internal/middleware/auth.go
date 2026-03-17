package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v5"
)

const (
	SessionName    = "gorundeck_session"
	SessionUserID  = "user_id"
	SessionRole    = "user_role"
	SessionUser    = "username"
)

// SessionStore is the global session store (set during server init).
var SessionStore sessions.Store

// RequireAuth is an Echo middleware that enforces a valid session.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		session, err := SessionStore.Get(c.Request(), SessionName)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		userID, ok := session.Values[SessionUserID]
		if !ok || userID == nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Populate context for downstream handlers
		c.Set(SessionUserID, userID)
		c.Set(SessionRole, session.Values[SessionRole])
		c.Set(SessionUser, session.Values[SessionUser])
		return next(c)
	}
}

// RequireAdmin enforces that the logged-in user has the admin role.
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		role, _ := c.Get(SessionRole).(string)
		if role != "admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin access required")
		}
		return next(c)
	}
}
