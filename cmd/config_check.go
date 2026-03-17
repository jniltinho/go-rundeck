package cmd

import (
	"fmt"

	"go-rundeck/config"
	"go-rundeck/internal/database"

	"github.com/spf13/cobra"
)

var configCheckCmd = &cobra.Command{
	Use:   "config-check",
	Short: "Validate configuration and test database connection",
	RunE:  runConfigCheck,
}

func init() {
	rootCmd.AddCommand(configCheckCmd)
}

func runConfigCheck(cmd *cobra.Command, args []string) error {
	fmt.Printf("Loading config from: %s\n", cfgFile)

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	fmt.Println("Config loaded successfully:")
	fmt.Printf("  App Name:  %s\n", cfg.App.Name)
	fmt.Printf("  Env:       %s\n", cfg.App.Env)
	fmt.Printf("  Port:      %d\n", cfg.App.Port)
	fmt.Printf("  DB Host:   %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("  DB Name:   %s\n", cfg.Database.Name)

	fmt.Println("\nTesting database connection…")
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	fmt.Println("Database connection OK.")
	return nil
}
