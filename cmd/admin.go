package cmd

import (
	"fmt"
	"strings"

	"go-rundeck/config"
	"go-rundeck/internal/database"
	"go-rundeck/internal/model"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var (
	addUserFlag       string
	sshHost           string
	sshPort           int
	sshUser           string
	sshPass           string
	sshCheckFlag      bool
	sshCheckDebugFlag bool
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin management commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addUserFlag != "" {
			return runAddUser(addUserFlag)
		}
		if sshCheckFlag {
			return runCheckSSH()
		}
		if sshCheckDebugFlag {
			return runCheckSSHDebug()
		}
		return cmd.Help()
	},
}

func init() {
	adminCmd.Flags().StringVar(&addUserFlag, "add-user", "", "Create a new admin user (format: email:password)")
	adminCmd.Flags().BoolVar(&sshCheckFlag, "check-ssh", false, "Test SSH connection to a host")
	adminCmd.Flags().BoolVar(&sshCheckDebugFlag, "check-ssh-debug", false, "Test SSH connection with verbose step-by-step debug output")
	adminCmd.Flags().StringVar(&sshHost, "host", "", "SSH host")
	adminCmd.Flags().IntVar(&sshPort, "port", 22, "SSH port")
	adminCmd.Flags().StringVar(&sshUser, "user", "root", "SSH user")
	adminCmd.Flags().StringVar(&sshPass, "pass", "", "SSH password")
	rootCmd.AddCommand(adminCmd)
}

func runAddUser(credentials string) error {
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format for --add-user. Expected format: email:password")
	}

	email := strings.TrimSpace(parts[0])
	password := parts[1]

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Extract a username from the email for the default username
	username := email
	if idx := strings.Index(email, "@"); idx != -1 {
		username = email[:idx]
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.Role("admin"), // Hardcoded to admin per the command intent
		Active:       true,
	}

	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user in database: %w", err)
	}

	fmt.Printf("Admin user '%s' (%s) created successfully.\n", username, email)
	return nil
}
