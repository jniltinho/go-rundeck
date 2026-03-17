package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/service"

	"github.com/labstack/echo/v4"
)

// ExecutionHandler handles execution routes including SSE log streaming.
type ExecutionHandler struct {
	execSvc    *service.ExecutionService
	projectSvc *service.ProjectService
}

// NewExecutionHandler creates a new ExecutionHandler.
func NewExecutionHandler(execSvc *service.ExecutionService, projectSvc *service.ProjectService) *ExecutionHandler {
	return &ExecutionHandler{execSvc: execSvc, projectSvc: projectSvc}
}

// List renders execution history for a project.
func (h *ExecutionHandler) List(c echo.Context) error {
	projectID, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := h.projectSvc.GetByID(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	execs, err := h.execSvc.List(projectID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Render(http.StatusOK, "executions/list.html", map[string]interface{}{
		"Title":       "Executions - " + project.Name,
		"Project":     project,
		"Executions":  execs,
		"Page":        page,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// Show renders execution detail.
func (h *ExecutionHandler) Show(c echo.Context) error {
	execID, err := parseID(c, "eid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid execution id")
	}
	exec, err := h.execSvc.GetByID(execID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "execution not found")
	}
	logs, err := h.execSvc.GetLogs(execID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Render(http.StatusOK, "executions/detail.html", map[string]interface{}{
		"Title":       fmt.Sprintf("Execution #%d", exec.ID),
		"Execution":   exec,
		"Logs":        logs,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// StreamLogs provides a Server-Sent Events stream of log entries.
func (h *ExecutionHandler) StreamLogs(c echo.Context) error {
	execID, err := parseID(c, "eid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid execution id")
	}

	exec, err := h.execSvc.GetByID(execID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "execution not found")
	}

	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Send existing logs first
	existingLogs, _ := h.execSvc.GetLogs(execID)
	for _, l := range existingLogs {
		fmt.Fprintf(w, "data: [%s][%s] %s\n\n", l.NodeName, l.LogLevel, l.Message)
		w.Flush()
	}

	// If execution is done, close stream
	if exec.Status != "running" {
		fmt.Fprintf(w, "event: done\ndata: execution finished\n\n")
		w.Flush()
		return nil
	}

	// Subscribe to live log events
	ch := h.execSvc.Subscribe(execID)
	defer h.execSvc.Unsubscribe(execID, ch)

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-ch:
			if !ok {
				fmt.Fprintf(w, "event: done\ndata: stream closed\n\n")
				w.Flush()
				return nil
			}
			l := event.Log
			fmt.Fprintf(w, "data: [%s][%s] %s\n\n", l.NodeName, l.LogLevel, l.Message)
			w.Flush()
		}
	}
}

// Abort stops a running execution.
func (h *ExecutionHandler) Abort(c echo.Context) error {
	execID, err := parseID(c, "eid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid execution id")
	}
	if err := h.execSvc.Abort(execID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/executions/"+c.Param("eid"))
}
