package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"go-rundeck/config"
	"go-rundeck/internal/database"
	"go-rundeck/internal/model"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user interactively",
	RunE:  runUserCreate,
}

func init() {
	userCmd.AddCommand(userCreateCmd)
	rootCmd.AddCommand(userCmd)
}

func runUserCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Role (admin/operator/viewer) [admin]: ")
	roleStr, _ := reader.ReadString('\n')
	roleStr = strings.TrimSpace(roleStr)
	if roleStr == "" {
		roleStr = "admin"
	}

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("read password: %w", err)
	}
	password := string(passwordBytes)

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.Role(roleStr),
		Active:       true,
	}

	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	fmt.Printf("User '%s' created successfully (role: %s).\n", username, roleStr)
	return nil
}
