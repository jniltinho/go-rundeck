package cmd

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"go-rundeck/config"
	"go-rundeck/internal/database"
	"go-rundeck/internal/logger"
	"go-rundeck/internal/router"

	"github.com/spf13/cobra"
)

// embedded filesystem references set by main via SetEmbeds.
var (
	templatesFS embed.FS
	staticFS    embed.FS
)

// SetEmbeds injects the embedded file systems from the main package.
func SetEmbeds(tFS, sFS embed.FS) {
	templatesFS = tFS
	staticFS = sFS
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Go-Rundeck HTTP server",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize logger
	logger.Init(cfg)

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	slog.Info("Starting server",
		slog.String("app", cfg.Server.Name),
		slog.String("addr", cfg.Server.Addr()),
		slog.String("env", cfg.Server.Env),
	)

	timeout := cfg.Server.SessionTimeout
	if timeout <= 0 {
		timeout = 60
	}
	sshTimeout := cfg.SSH.ConnectTimeout
	if sshTimeout <= 0 {
		sshTimeout = 10
	}
	e := router.Setup(db, templatesFS, staticFS, cfg.Server.SessionSecret, timeout, cfg.Server.SSLEnabled, appVersion, sshTimeout)

	if cfg.Server.SSLEnabled {
		if cfg.Server.SSLCert == "" || cfg.Server.SSLKey == "" {
			slog.Error("SSL enabled but cert or key file not provided")
			return fmt.Errorf("server.ssl_cert and server.ssl_key must be set when ssl_enable is true")
		}
		slog.Info("SSL enabled", "cert", cfg.Server.SSLCert, "key", cfg.Server.SSLKey)
		server := &http.Server{Addr: cfg.Server.Addr(), Handler: e}
		return server.ListenAndServeTLS(cfg.Server.SSLCert, cfg.Server.SSLKey)
	}

	return e.Start(cfg.Server.Addr())
}
