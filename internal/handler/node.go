package handler

import (
	"net/http"
	"strconv"

	"go-rundeck/internal/middleware"
	"go-rundeck/internal/model"
	"go-rundeck/internal/repository"
	"go-rundeck/internal/service"

	"github.com/labstack/echo/v5"
)

type NodeHandler struct {
	nodeRepo   *repository.NodeRepository
	projectSvc *service.ProjectService
	sshSvc     *service.SSHService
	keySvc     *service.KeyService
}

// NewNodeHandler creates a new NodeHandler.
func NewNodeHandler(nodeRepo *repository.NodeRepository, projectSvc *service.ProjectService, sshSvc *service.SSHService, keySvc *service.KeyService) *NodeHandler {
	return &NodeHandler{
		nodeRepo:   nodeRepo,
		projectSvc: projectSvc,
		sshSvc:     sshSvc,
		keySvc:     keySvc,
	}
}

// List renders the node list for a project.
func (h *NodeHandler) List(c *echo.Context) error {
	projectID, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}
	project, err := h.projectSvc.GetByID(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "project not found")
	}
	nodes, err := h.nodeRepo.ListByProject(projectID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	
	keys, err := h.keySvc.ListSystemKeys()
	if err != nil {
		keys = make([]model.KeyStorage, 0) // fall back to empty if we can't load them
	}

	return c.Render(http.StatusOK, "nodes/list.html", map[string]interface{}{
		"Title":       "Nodes - " + project.Name,
		"Project":     project,
		"Nodes":       nodes,
		"Keys":        keys,
		"CurrentUser": c.Get(middleware.SessionUser),
		"Role":        c.Get(middleware.SessionRole),
	})
}

// Show renders a single node detail.
func (h *NodeHandler) Show(c *echo.Context) error {
	nodeID, err := parseID(c, "nid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node id")
	}
	node, err := h.nodeRepo.GetByID(nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "node not found")
	}
	keys, err := h.keySvc.ListSystemKeys()
	if err != nil {
		keys = make([]model.KeyStorage, 0)
	}
	var selectedKeyID uint
	if node.KeyID != nil {
		selectedKeyID = *node.KeyID
	}
	return c.Render(http.StatusOK, "nodes/detail.html", map[string]interface{}{
		"Title":         node.Name,
		"Node":          node,
		"Keys":          keys,
		"SelectedKeyID": selectedKeyID,
		"CurrentUser":   c.Get(middleware.SessionUser),
		"Role":          c.Get(middleware.SessionRole),
	})
}

// Create handles node creation form.
func (h *NodeHandler) Create(c *echo.Context) error {
	projectID, err := parseID(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	sshPort, _ := strconv.Atoi(c.FormValue("ssh_port"))
	if sshPort == 0 {
		sshPort = 22
	}

	node := &model.Node{
		ProjectID:   projectID,
		Name:        c.FormValue("name"),
		Hostname:    c.FormValue("hostname"),
		SSHPort:     sshPort,
		SSHUser:     c.FormValue("ssh_user"),
		AuthType:    model.AuthType(c.FormValue("auth_type")),
		Tags:        c.FormValue("tags"),
		Description: c.FormValue("description"),
		OSFamily:    c.FormValue("os_family"),
		Active:      true,
	}

	keyIDStr := c.FormValue("key_id")
	if keyIDStr != "" {
		kid, _ := strconv.ParseUint(keyIDStr, 10, 32)
		kidu := uint(kid)
		node.KeyID = &kidu
	}

	if err := h.nodeRepo.Create(node); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects/"+c.Param("id")+"/nodes")
}

// Update handles node update form.
func (h *NodeHandler) Update(c *echo.Context) error {
	nodeID, err := parseID(c, "nid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node id")
	}
	node, err := h.nodeRepo.GetByID(nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "node not found")
	}

	sshPort, _ := strconv.Atoi(c.FormValue("ssh_port"))
	if sshPort > 0 {
		node.SSHPort = sshPort
	}
	if v := c.FormValue("name"); v != "" {
		node.Name = v
	}
	if v := c.FormValue("hostname"); v != "" {
		node.Hostname = v
	}
	if v := c.FormValue("ssh_user"); v != "" {
		node.SSHUser = v
	}
	node.Tags = c.FormValue("tags")
	node.Description = c.FormValue("description")
	node.OSFamily = c.FormValue("os_family")
	if v := c.FormValue("auth_type"); v != "" {
		node.AuthType = model.AuthType(v)
	}

	keyIDStr := c.FormValue("key_id")
	if keyIDStr == "" || keyIDStr == "0" {
		node.KeyID = nil
	} else {
		kid, _ := strconv.ParseUint(keyIDStr, 10, 32)
		kidu := uint(kid)
		node.KeyID = &kidu
	}

	if err := h.nodeRepo.Update(node); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects/"+c.Param("id")+"/nodes")
}

// CheckSSH tests SSH connectivity for a node and returns JSON.
func (h *NodeHandler) CheckSSH(c *echo.Context) error {
	nodeID, err := parseID(c, "nid")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid node id"})
	}
	node, err := h.nodeRepo.GetByID(nodeID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "node not found"})
	}

	if node.AuthType == "password" {
		password := ""
		if node.KeyID != nil {
			password, _ = h.keySvc.GetDecryptedContent(*node.KeyID)
		}
		_, err = h.sshSvc.RunCommandWithPassword(node.Hostname, node.SSHPort, node.SSHUser, password, "echo ok")
	} else {
		if node.KeyID == nil {
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "no key configured"})
		}
		pemKey, keyErr := h.keySvc.GetDecryptedContent(*node.KeyID)
		if keyErr != nil || pemKey == "" {
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "could not load key"})
		}
		_, err = h.sshSvc.RunCommandWithKey(node.Hostname, node.SSHPort, node.SSHUser, []byte(pemKey), "echo ok")
	}

	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
}

// Delete soft-deletes a node.
func (h *NodeHandler) Delete(c *echo.Context) error {
	nodeID, err := parseID(c, "nid")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node id")
	}
	if err := h.nodeRepo.Delete(nodeID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/projects/"+c.Param("id")+"/nodes")
}
