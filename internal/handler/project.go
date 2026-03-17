package handler

import (
	"net/http"
	"strconv"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/service"

	"github.com/labstack/echo/v4"
)

// ProjectHandler handles project CRUD routes.
type ProjectHandler struct {
	svc *service.ProjectService
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List renders the projects list page.
func (h *ProjectHandler) List(c echo.Context) error {
	projects, err := h.svc.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Render(http.StatusOK, "projects/list.html", map[string]interface{}{
		"Title":       "Projects",
		"Projects":    projects,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// Show renders the project detail page.
func (h *ProjectHandler) Show(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	p, err := h.svc.GetByID(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	return c.Render(http.StatusOK, "projects/detail.html", map[string]interface{}{
		"Title":       p.Name,
		"Project":     p,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// Create handles project creation form submission.
func (h *ProjectHandler) Create(c echo.Context) error {
	userID := c.Get(middleware.SessionUserID).(uint)
	name := c.FormValue("name")
	description := c.FormValue("description")
	tags := c.FormValue("tags")

	_, err := h.svc.Create(name, description, tags, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects")
}

// Update handles project update form submission.
func (h *ProjectHandler) Update(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	name := c.FormValue("name")
	description := c.FormValue("description")
	tags := c.FormValue("tags")

	if _, err := h.svc.Update(id, name, description, tags); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects/"+strconv.Itoa(int(id)))
}

// Delete soft-deletes a project.
func (h *ProjectHandler) Delete(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	if err := h.svc.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects")
}

// parseID extracts and converts a URL param to uint.
func parseID(c echo.Context, param string) (uint, error) {
	raw := c.Param(param)
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}
