package cmd

import (
	"embed"
	"fmt"
	"log/slog"

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
		slog.String("app", cfg.App.Name),
		slog.String("addr", cfg.App.Addr()),
		slog.String("env", cfg.App.Env),
	)

	e := router.Setup(db, templatesFS, staticFS, cfg.App.Secret)
	return e.Start(cfg.App.Addr())
}
