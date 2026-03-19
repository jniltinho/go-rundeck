package handler

import (
	"net/http"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/model"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler handles login/logout.
type AuthHandler struct {
	db      *gorm.DB
	version string
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(db *gorm.DB, version string) *AuthHandler {
	return &AuthHandler{db: db, version: version}
}

// ShowLogin renders the login page.
func (h *AuthHandler) ShowLogin(c *echo.Context) error {
	return c.Render(http.StatusOK, "auth/login.html", map[string]interface{}{
		"Title":   "Login",
		"Error":   "",
		"Version": h.version,
	})
}

// Login handles form submission and creates a session.
func (h *AuthHandler) Login(c *echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	var user model.User
	if err := h.db.Where("username = ? AND active = ?", username, true).First(&user).Error; err != nil {
		return c.Render(http.StatusUnauthorized, "auth/login.html", map[string]interface{}{
			"Title":   "Login",
			"Error":   "Invalid username or password",
			"Version": h.version,
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return c.Render(http.StatusUnauthorized, "auth/login.html", map[string]interface{}{
			"Title":   "Login",
			"Error":   "Invalid username or password",
			"Version": h.version,
		})
	}

	session, _ := middleware.SessionStore.Get(c.Request(), middleware.SessionName)
	session.Values[middleware.SessionUserID] = user.ID
	session.Values[middleware.SessionRole] = string(user.Role)
	session.Values[middleware.SessionUser] = user.Username
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session")
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// Logout destroys the session.
func (h *AuthHandler) Logout(c *echo.Context) error {
	session, _ := middleware.SessionStore.Get(c.Request(), middleware.SessionName)
	session.Options.MaxAge = -1
	_ = session.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusSeeOther, "/login")
}
