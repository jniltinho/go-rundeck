package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"go-rundeck/internal/model"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// List shows all users
func (h *UserHandler) List(c *echo.Context) error {
	var users []model.User
	if err := h.db.Find(&users).Error; err != nil {
		slog.Error("failed to list users", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load users")
	}

	return c.Render(http.StatusOK, "user/list.html", map[string]interface{}{
		"Title": "Users",
		"Users": users,
		"Role":  c.Get("user_role"),
	})
}

// Create handles creating a new user
func (h *UserHandler) Create(c *echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	email := c.FormValue("email")
	role := c.FormValue("role")

	if username == "" || password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Username and password are required")
	}
	if len(password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "Password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process password")
	}

	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Email:        email,
		Role:         model.Role(role),
		Active:       true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		slog.Error("failed to create user", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.Redirect(http.StatusSeeOther, "/users")
}

// Update handles editing a user's details
func (h *UserHandler) Update(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	if v := c.FormValue("username"); v != "" {
		user.Username = v
	}
	user.Email = c.FormValue("email")
	if v := c.FormValue("role"); v != "" {
		user.Role = model.Role(v)
	}
	user.Active = c.FormValue("active") == "1"

	if pw := c.FormValue("password"); pw != "" {
		if len(pw) < 8 {
			return echo.NewHTTPError(http.StatusBadRequest, "Password must be at least 8 characters")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to process password")
		}
		user.PasswordHash = string(hash)
	}

	if err := h.db.Save(&user).Error; err != nil {
		slog.Error("failed to update user", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
	}

	return c.Redirect(http.StatusSeeOther, "/users")
}

// Delete handles deleting or deactivating a user
func (h *UserHandler) Delete(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	// Permanently delete the user from the database
	if err := h.db.Unscoped().Delete(&model.User{}, id).Error; err != nil {
		slog.Error("failed to delete user", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete user")
	}

	return c.Redirect(http.StatusSeeOther, "/users")
}
