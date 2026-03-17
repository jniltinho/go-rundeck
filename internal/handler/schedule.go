package handler

import (
	"net/http"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/model"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// ScheduleHandler handles schedule CRUD routes.
type ScheduleHandler struct {
	db *gorm.DB
}

// NewScheduleHandler creates a new ScheduleHandler.
func NewScheduleHandler(db *gorm.DB) *ScheduleHandler {
	return &ScheduleHandler{db: db}
}

// ListByJob returns schedules for a specific job as JSON.
func (h *ScheduleHandler) ListByJob(c *echo.Context) error {
	jobID, err := parseID(c, "jid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job id")
	}
	var schedules []model.Schedule
	if err := h.db.Where("job_id = ?", jobID).Find(&schedules).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, schedules)
}

// Create creates a schedule for a job.
func (h *ScheduleHandler) Create(c *echo.Context) error {
	jobID, err := parseID(c, "jid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job id")
	}
	cronExpr := c.FormValue("cron_expr")
	if cronExpr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "cron_expr is required")
	}
	sched := &model.Schedule{
		JobID:    jobID,
		CronExpr: cronExpr,
		Enabled:  true,
	}
	if err := h.db.Create(sched).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, sched)
}

// Delete removes a schedule.
func (h *ScheduleHandler) Delete(c *echo.Context) error {
	id, err := parseID(c, "sid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid schedule id")
	}
	if err := h.db.Delete(&model.Schedule{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	_ = c.Get(middleware.SessionUser) // ensure middleware import is used
	return c.NoContent(http.StatusNoContent)
}
