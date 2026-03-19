package router

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"go-rundeck/internal/handler"
	mw "go-rundeck/internal/middleware"
	"go-rundeck/internal/repository"
	"go-rundeck/internal/service"

	"github.com/gorilla/sessions"
	echomw "github.com/labstack/echo/v5/middleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// TemplateRenderer implements echo.Renderer using html/template with embedded FS.
type TemplateRenderer struct {
	templates *template.Template
}

// Render writes the named template to w.
func (t *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Setup configures and returns an Echo instance with all routes registered.
func Setup(
	db *gorm.DB,
	templatesFS embed.FS,
	staticFS embed.FS,
	secret string,
	sessionTimeout int,
	sslEnabled bool,
	version string,
	sshConnectTimeout int,
) *echo.Echo {
	e := echo.New()
	e.Logger = slog.Default()

	// ── Template renderer ────────────────────────────────────────────────────
	tmpl := parseTemplates(templatesFS)
	e.Renderer = &TemplateRenderer{templates: tmpl}

	// ── Session store ────────────────────────────────────────────────────────
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionTimeout * 60,
		HttpOnly: true,
		Secure:   sslEnabled,
		SameSite: http.SameSiteLaxMode,
	}
	mw.SessionStore = store
	mw.SessionTimeout = sessionTimeout

	// ── Global middleware ─────────────────────────────────────────────────────
	e.Use(echomw.Recover())
	e.Use(mw.RequestLogger())
	e.Use(mw.CORS())

	// ── Static assets from embedded FS ───────────────────────────────────────
	sub, _ := fs.Sub(staticFS, "web/static")
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(sub)))))

	// ── Repositories / Services / Handlers ───────────────────────────────────
	projectRepo := repository.NewProjectRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	jobRepo := repository.NewJobRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	keyRepo, err := repository.NewKeyRepository(db, secret)
	if err != nil {
		slog.Error("failed to init key repository", "error", err)
	}

	projectSvc := service.NewProjectService(projectRepo)
	sshSvc := service.NewSSHService(sshConnectTimeout)
	execSvc := service.NewExecutionService(execRepo)
	keySvc := service.NewKeyService(keyRepo)
	jobSvc := service.NewJobService(jobRepo, nodeRepo, execSvc, sshSvc, keySvc)

	authH := handler.NewAuthHandler(db, version)
	dashH := handler.NewDashboardHandler(projectSvc, execSvc, nodeRepo, jobRepo, execRepo)
	projectH := handler.NewProjectHandler(projectSvc)
	nodeH := handler.NewNodeHandler(nodeRepo, projectSvc, sshSvc, keySvc)
	jobH := handler.NewJobHandler(jobSvc, projectSvc)
	execH := handler.NewExecutionHandler(execSvc, projectSvc)
	schedH := handler.NewScheduleHandler(db)
	userH := handler.NewUserHandler(db)
	keyH := handler.NewKeyHandler(keySvc)

	// ── Public routes ─────────────────────────────────────────────────────────
	e.GET("/login", authH.ShowLogin)
	e.POST("/login", authH.Login)
	e.POST("/logout", authH.Logout)

	// ── Protected routes ─────────────────────────────────────────────────────
	protected := e.Group("", mw.RequireAuth)

	protected.GET("/", dashH.Index)

	// Projects
	protected.GET("/projects", projectH.List)
	protected.POST("/projects", projectH.Create)
	protected.GET("/projects/:id", projectH.Show)
	protected.POST("/projects/:id", projectH.Update)
	protected.POST("/projects/:id/delete", projectH.Delete)

	// Nodes
	protected.GET("/projects/:id/nodes", nodeH.List)
	protected.POST("/projects/:id/nodes", nodeH.Create)
	protected.GET("/projects/:id/nodes/:nid", nodeH.Show)
	protected.POST("/projects/:id/nodes/:nid", nodeH.Update)
	protected.POST("/projects/:id/nodes/:nid/delete", nodeH.Delete)
	protected.POST("/projects/:id/nodes/:nid/check-ssh", nodeH.CheckSSH)
	protected.POST("/projects/:id/nodes/:nid/toggle-active", nodeH.ToggleActive)

	// Jobs
	protected.GET("/projects/:id/jobs", jobH.List)
	protected.GET("/projects/:id/jobs/new", jobH.ShowCreate)
	protected.POST("/projects/:id/jobs", jobH.Create)
	protected.GET("/projects/:id/jobs/:jid", jobH.Show)
	protected.POST("/projects/:id/jobs/:jid", jobH.Update)
	protected.POST("/projects/:id/jobs/:jid/delete", jobH.Delete)
	protected.POST("/projects/:id/jobs/:jid/run", jobH.Run)

	// Schedules
	protected.GET("/projects/:id/jobs/:jid/schedules", schedH.ListByJob)
	protected.POST("/projects/:id/jobs/:jid/schedules", schedH.Create)
	protected.POST("/projects/:id/jobs/:jid/schedules/:sid/toggle", schedH.Toggle)
	protected.POST("/projects/:id/jobs/:jid/schedules/:sid/delete", schedH.Delete)

	// Executions
	protected.GET("/projects/:id/executions", execH.List)
	protected.GET("/executions/:eid", execH.Show)
	protected.GET("/executions/:eid/log", execH.StreamLogs)
	protected.POST("/executions/:eid/abort", execH.Abort)
	protected.POST("/executions/:eid/delete", execH.Delete)

	// Users (Admin Only)
	adminGrp := e.Group("/users", mw.RequireAuth, mw.RequireAdmin)
	adminGrp.GET("", userH.List)
	adminGrp.POST("", userH.Create)
	adminGrp.POST("/:id", userH.Update)
	adminGrp.POST("/:id/delete", userH.Delete)

	// Key Storage
	protected.GET("/keys", keyH.ListSystemKeys)
	protected.POST("/keys", keyH.Create)
	protected.POST("/keys/:id", keyH.Update)
	protected.POST("/keys/:id/delete", keyH.Delete)

	return e
}

// parseTemplates loads all *.html files from the embedded FS into a single Template set.
func parseTemplates(fsys embed.FS) *template.Template {
	tmpl := template.New("").Funcs(templateFuncMap())

	err := fs.WalkDir(fsys, "web/templates", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(filePath, ".html") {
			return nil
		}

		content, readErr := fsys.ReadFile(filePath)
		if readErr != nil {
			return readErr
		}

		// Use path relative to web/templates as template name
		name := strings.TrimPrefix(filePath, "web/templates/")
		t := tmpl.New(name)
		if _, parseErr := t.Parse(string(content)); parseErr != nil {
			return parseErr
		}
		return nil
	})
	if err != nil {
		panic("failed to parse templates: " + err.Error())
	}
	return tmpl
}
