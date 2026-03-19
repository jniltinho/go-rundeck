package cmd

import (
	"fmt"

	"go-rundeck/config"
	"go-rundeck/internal/service"
)

func runCheckSSHDebug() error {
	if sshHost == "" {
		return fmt.Errorf("--host is required")
	}
	if sshPass == "" {
		return fmt.Errorf("--pass is required")
	}

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

func runCheckSSH() error {
	if sshHost == "" {
		return fmt.Errorf("--host is required")
	}
	if sshPass == "" {
		return fmt.Errorf("--pass is required")
	}

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
