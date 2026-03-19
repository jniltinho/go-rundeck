package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/model"
	"go-rundeck/internal/service"

	"github.com/labstack/echo/v5"
)

type KeyHandler struct {
	svc *service.KeyService
}

func NewKeyHandler(svc *service.KeyService) *KeyHandler {
	return &KeyHandler{svc: svc}
}

// ListSystemKeys renders the keys manager page.
func (h *KeyHandler) ListSystemKeys(c *echo.Context) error {
	keys, err := h.svc.ListSystemKeys()
	if err != nil {
		slog.Error("failed to list keys", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load keys")
	}

	return c.Render(http.StatusOK, "keys/list.html", map[string]interface{}{
		"Title":       "Key Storage",
		"Keys":        keys,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// Create handles form submission to create a key/password.
func (h *KeyHandler) Create(c *echo.Context) error {
	userID := c.Get(middleware.SessionUserID).(uint)
	name := c.FormValue("name")
	keyType := c.FormValue("type")
	content := c.FormValue("content")
	description := c.FormValue("description")

	var kType model.KeyType
	if keyType == "private_key" {
		kType = model.KeyTypePrivateKey
	} else {
		kType = model.KeyTypePassword
	}

	if _, err := h.svc.Create(name, kType, content, description, nil, userID); err != nil {
		slog.Error("failed to create key", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to store key")
	}

	return c.Redirect(http.StatusSeeOther, "/keys")
}

// Update handles editing a key's description and optionally its content.
func (h *KeyHandler) Update(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid key ID")
	}

	name := c.FormValue("name")
	keyType := c.FormValue("type")
	description := c.FormValue("description")
	newContent := c.FormValue("content")

	var kType model.KeyType
	if keyType == "private_key" {
		kType = model.KeyTypePrivateKey
	} else {
		kType = model.KeyTypePassword
	}

	if err := h.svc.Update(uint(id), name, kType, description, newContent); err != nil {
		slog.Error("failed to update key", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update key")
	}

	return c.Redirect(http.StatusSeeOther, "/keys")
}

// Delete handles removing a key.
func (h *KeyHandler) Delete(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid key ID")
	}

	if err := h.svc.Delete(uint(id)); err != nil {
		slog.Error("failed to delete key", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete key")
	}

	return c.Redirect(http.StatusSeeOther, "/keys")
}
