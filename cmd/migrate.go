package cmd

import (
	"fmt"
	"log"

	"go-rundeck/config"
	"go-rundeck/internal/database"
	"go-rundeck/internal/model"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations (AutoMigrate)",
	RunE:  runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	log.Println("Running AutoMigrate…")

	if err := db.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.KeyStorage{},
		&model.Node{},
		&model.Job{},
		&model.JobStep{},
		&model.Execution{},
		&model.ExecutionLog{},
		&model.Schedule{},
		&model.JobOption{},
		&model.ExecutionOption{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	log.Println("Migration completed successfully.")
	return nil
}
