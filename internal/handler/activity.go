package handler

import (
	"net/http"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/repository"
	"go-rundeck/internal/service"

	"github.com/labstack/echo/v4"
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
func (h *DashboardHandler) Index(c echo.Context) error {
	projectCount, _ := h.projectSvc.Count()
	runningCount, _ := h.execSvc.CountRunning()
	recentActivity, _ := h.execSvc.RecentActivity(10)

	return c.Render(http.StatusOK, "dashboard/index.html", map[string]interface{}{
		"Title":          "Dashboard",
		"ProjectCount":   projectCount,
		"RunningCount":   runningCount,
		"RecentActivity": recentActivity,
		"CurrentUser":    c.Get(middleware.SessionUser),
		"Role":           c.Get(middleware.SessionRole),
	})
}
