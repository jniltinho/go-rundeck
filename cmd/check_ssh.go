package cmd

import (
	"fmt"

	"go-rundeck/config"
	"go-rundeck/internal/service"

	"github.com/spf13/cobra"
)

var (
	sshHost  string
	sshPort  int
	sshUser  string
	sshPass  string
	sshDebug bool
)

var checkSSHCmd = &cobra.Command{
	Use:   "check-ssh",
	Short: "Test SSH connectivity to a host",
	RunE: func(cmd *cobra.Command, args []string) error {
		if sshDebug {
			return runCheckSSHDebug()
		}
		return runCheckSSH()
	},
}

func init() {
	checkSSHCmd.Flags().StringVar(&sshHost, "host", "", "SSH host (required)")
	checkSSHCmd.Flags().IntVar(&sshPort, "port", 22, "SSH port")
	checkSSHCmd.Flags().StringVar(&sshUser, "user", "root", "SSH user")
	checkSSHCmd.Flags().StringVar(&sshPass, "pass", "", "SSH password (required)")
	checkSSHCmd.Flags().BoolVar(&sshDebug, "debug", false, "Verbose step-by-step debug output")
	checkSSHCmd.MarkFlagRequired("host")
	checkSSHCmd.MarkFlagRequired("pass")
	rootCmd.AddCommand(checkSSHCmd)
}

func runCheckSSH() error {
	fmt.Printf("Testing SSH connection to %s@%s:%d ...\n", sshUser, sshHost, sshPort)

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	sshSvc := service.NewSSHService(cfg.SSH.ConnectTimeout)
	result, err := sshSvc.RunCommandWithPassword(sshHost, sshPort, sshUser, sshPass, "hostname && whoami && date")
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	fmt.Println("─── stdout ───────────────────────────────")
	fmt.Print(result.Stdout)
	if result.Stderr != "" {
		fmt.Println("─── stderr ───────────────────────────────")
		fmt.Print(result.Stderr)
	}
	fmt.Printf("─── exit code: %d ────────────────────────\n", result.ExitCode)
	return nil
}

func runCheckSSHDebug() error {
	logf := func(format string, args ...any) {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}

	logf("starting debug SSH test")

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	sshSvc := service.NewSSHService(cfg.SSH.ConnectTimeout)
	result, err := sshSvc.RunCommandWithPasswordDebug(sshHost, sshPort, sshUser, sshPass, "hostname && whoami && date", logf)
	if err != nil {
		return fmt.Errorf("SSH failed: %w", err)
	}

	fmt.Println("─── stdout ───────────────────────────────")
	fmt.Print(result.Stdout)
	if result.Stderr != "" {
		fmt.Println("─── stderr ───────────────────────────────")
		fmt.Print(result.Stderr)
	}
	fmt.Printf("─── exit code: %d ────────────────────────\n", result.ExitCode)
	return nil
}
