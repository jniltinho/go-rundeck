package handler

import (
	"net/http"
	"strconv"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/model"
	"go-rundeck/internal/service"

	"github.com/labstack/echo/v5"
)

// JobHandler handles job CRUD and run routes.
type JobHandler struct {
	jobSvc     *service.JobService
	projectSvc *service.ProjectService
}

// NewJobHandler creates a new JobHandler.
func NewJobHandler(jobSvc *service.JobService, projectSvc *service.ProjectService) *JobHandler {
	return &JobHandler{jobSvc: jobSvc, projectSvc: projectSvc}
}

// List renders the job list for a project.
func (h *JobHandler) List(c *echo.Context) error {
	projectID, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := h.projectSvc.GetByID(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	jobs, err := h.jobSvc.ListByProject(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Render(http.StatusOK, "jobs/list.html", map[string]interface{}{
		"Title":       "Jobs - " + project.Name,
		"Project":     project,
		"Jobs":        jobs,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// ShowCreate renders the job creation form.
func (h *JobHandler) ShowCreate(c *echo.Context) error {
	projectID, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := h.projectSvc.GetByID(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	return c.Render(http.StatusOK, "jobs/create.html", map[string]interface{}{
		"Title":       "New Job",
		"Project":     project,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// Show renders a job detail page.
func (h *JobHandler) Show(c *echo.Context) error {
	jobID, err := parseID(c, "jid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job id")
	}
	job, err := h.jobSvc.GetByID(jobID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}
	return c.Render(http.StatusOK, "jobs/detail.html", map[string]interface{}{
		"Title":       job.Name,
		"Job":         job,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// Create handles the job creation form submission.
func (h *JobHandler) Create(c *echo.Context) error {
	projectID, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	userID := c.Get(middleware.SessionUserID).(uint)

	timeoutSec, _ := strconv.Atoi(c.FormValue("timeout_sec"))
	if timeoutSec == 0 {
		timeoutSec = 300
	}

	job := &model.Job{
		ProjectID:    projectID,
		Name:         c.FormValue("name"),
		Description:  c.FormValue("description"),
		NodeFilter:   c.FormValue("node_filter"),
		ExecStrategy: model.ExecStrategy(c.FormValue("exec_strategy")),
		OnError:      model.OnError(c.FormValue("on_error")),
		TimeoutSec:   timeoutSec,
		CreatedBy:    userID,
	}
	if job.ExecStrategy == "" {
		job.ExecStrategy = model.ExecStrategySequential
	}
	if job.OnError == "" {
		job.OnError = model.OnErrorStop
	}

	created, err := h.jobSvc.Create(job)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects/"+c.Param("id")+"/jobs/"+strconv.Itoa(int(created.ID)))
}

// Update handles job update form.
func (h *JobHandler) Update(c *echo.Context) error {
	jobID, err := parseID(c, "jid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job id")
	}
	job, err := h.jobSvc.GetByID(jobID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	if v := c.FormValue("name"); v != "" {
		job.Name = v
	}
	job.Description = c.FormValue("description")
	job.NodeFilter = c.FormValue("node_filter")
	if v := c.FormValue("exec_strategy"); v != "" {
		job.ExecStrategy = model.ExecStrategy(v)
	}
	if v := c.FormValue("on_error"); v != "" {
		job.OnError = model.OnError(v)
	}
	if ts, err2 := strconv.Atoi(c.FormValue("timeout_sec")); err2 == nil && ts > 0 {
		job.TimeoutSec = ts
	}

	if err := h.jobSvc.Update(job); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects/"+c.Param("id")+"/jobs/"+c.Param("jid"))
}

// Delete soft-deletes a job.
func (h *JobHandler) Delete(c *echo.Context) error {
	jobID, err := parseID(c, "jid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job id")
	}
	if err := h.jobSvc.Delete(jobID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects/"+c.Param("id")+"/jobs")
}

// Run triggers a job execution.
func (h *JobHandler) Run(c *echo.Context) error {
	jobID, err := parseID(c, "jid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job id")
	}
	userIDVal := c.Get(middleware.SessionUserID)
	var userID *uint
	if uid, ok := userIDVal.(uint); ok {
		userID = &uid
	}

	exec, err := h.jobSvc.Run(jobID, userID, model.TriggerTypeManual)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/executions/"+strconv.Itoa(int(exec.ID)))
}
