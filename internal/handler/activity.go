package handler

import (
	"net/http"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/repository"
	"go-rundeck/internal/service"

	"github.com/labstack/echo/v5"
)

// DashboardHandler handles the main dashboard.
type DashboardHandler struct {
	projectSvc  *service.ProjectService
	execSvc     *service.ExecutionService
	nodeRepo    *repository.NodeRepository
	jobRepo     *repository.JobRepository
	execRepo    *repository.ExecutionRepository
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	projectSvc *service.ProjectService,
	execSvc *service.ExecutionService,
	nodeRepo *repository.NodeRepository,
	jobRepo *repository.JobRepository,
	execRepo *repository.ExecutionRepository,
) *DashboardHandler {
	return &DashboardHandler{
		projectSvc: projectSvc,
		execSvc:    execSvc,
		nodeRepo:   nodeRepo,
		jobRepo:    jobRepo,
		execRepo:   execRepo,
	}
}

// Index renders the dashboard.
func (h *DashboardHandler) Index(c *echo.Context) error {
	projectCount, _ := h.projectSvc.Count()
	projects, _ := h.projectSvc.List()
	lastDayCount, _ := h.execSvc.CountLastDay()
	lastDayFailed, _ := h.execSvc.CountFailedLastDay()

	return c.Render(http.StatusOK, "dashboard/index.html", map[string]interface{}{
		"Title":         "Home",
		"ProjectCount":  projectCount,
		"Projects":      projects,
		"LastDayCount":  lastDayCount,
		"LastDayFailed": lastDayFailed,
		"CurrentUser":   c.Get(middleware.SessionUser),
		"Role":          c.Get(middleware.SessionRole),
	})
}
