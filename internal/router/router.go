package router

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"go-rundeck/internal/handler"
	mw "go-rundeck/internal/middleware"
	"go-rundeck/internal/repository"
	"go-rundeck/internal/service"

	"github.com/gorilla/sessions"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// TemplateRenderer implements echo.Renderer using html/template with embedded FS.
type TemplateRenderer struct {
	templates *template.Template
}

// Render writes the named template to w.
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Setup configures and returns an Echo instance with all routes registered.
func Setup(
	db *gorm.DB,
	templatesFS embed.FS,
	staticFS embed.FS,
	secret string,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// ── Template renderer ────────────────────────────────────────────────────
	tmpl := parseTemplates(templatesFS)
	e.Renderer = &TemplateRenderer{templates: tmpl}

	// ── Session store ────────────────────────────────────────────────────────
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	mw.SessionStore = store

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

	projectSvc := service.NewProjectService(projectRepo)
	sshSvc := service.NewSSHService(10)
	execSvc := service.NewExecutionService(execRepo)
	jobSvc := service.NewJobService(jobRepo, nodeRepo, execSvc, sshSvc)

	authH := handler.NewAuthHandler(db)
	dashH := handler.NewDashboardHandler(projectSvc, execSvc, nodeRepo, jobRepo, execRepo)
	projectH := handler.NewProjectHandler(projectSvc)
	nodeH := handler.NewNodeHandler(nodeRepo, projectSvc, sshSvc)
	jobH := handler.NewJobHandler(jobSvc, projectSvc)
	execH := handler.NewExecutionHandler(execSvc, projectSvc)
	schedH := handler.NewScheduleHandler(db)

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
	protected.POST("/projects/:id/jobs/:jid/schedules/:sid/delete", schedH.Delete)

	// Executions
	protected.GET("/projects/:id/executions", execH.List)
	protected.GET("/executions/:eid", execH.Show)
	protected.GET("/executions/:eid/log", execH.StreamLogs)
	protected.POST("/executions/:eid/abort", execH.Abort)

	return e
}

// parseTemplates loads all *.html files from the embedded FS into a single Template set.
func parseTemplates(fsys embed.FS) *template.Template {
	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"base":  path.Base,
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
	}

	tmpl := template.New("").Funcs(funcMap)

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
