package cmd

import (
	"fmt"
	"strings"
	"time"

	"go-rundeck/config"
	"go-rundeck/internal/service"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
			runCheckSSHDebug()
			return nil
		}
		runCheckSSH()
		return nil
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
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	start := time.Now()
	sshSvc := service.NewSSHService(cfg.SSH.ConnectTimeout)
	result, err := sshSvc.RunCommandWithPassword(sshHost, sshPort, sshUser, sshPass, "hostname && whoami && date")
	elapsed := time.Since(start)

	t := table.NewWriter()
	t.Style().Options.SeparateRows = true

	if err != nil {
		t.SetTitle(text.FgRed.Sprint("SSH Connection Test — FAILED"))
		t.AppendRows([]table.Row{
			{"Host", fmt.Sprintf("%s:%d", sshHost, sshPort)},
			{"User", sshUser},
			{"Duration", elapsed.Round(time.Millisecond)},
			{"Error", err.Error()},
		})
		fmt.Println(t.Render())
		return err
	}

	t.SetTitle(text.FgGreen.Sprint("SSH Connection Test — SUCCESS"))
	t.AppendRows([]table.Row{
		{"Host", fmt.Sprintf("%s:%d", sshHost, sshPort)},
		{"User", sshUser},
		{"Duration", elapsed.Round(time.Millisecond)},
		{"Exit Code", result.ExitCode},
		{"Output", strings.TrimSpace(result.Stdout)},
	})
	if result.Stderr != "" {
		t.AppendRow(table.Row{"Stderr", strings.TrimSpace(result.Stderr)})
	}
	fmt.Println(t.Render())
	return nil
}

func runCheckSSHDebug() error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var logLines []string
	logf := func(format string, args ...any) {
		logLines = append(logLines, fmt.Sprintf(format, args...))
	}

	logf("starting debug SSH test to %s@%s:%d", sshUser, sshHost, sshPort)

	start := time.Now()
	sshSvc := service.NewSSHService(cfg.SSH.ConnectTimeout)
	result, err := sshSvc.RunCommandWithPasswordDebug(sshHost, sshPort, sshUser, sshPass, "hostname && whoami && date", logf)
	elapsed := time.Since(start)

	t := table.NewWriter()
	t.Style().Options.SeparateRows = true

	if err != nil {
		t.SetTitle(text.FgRed.Sprint("SSH Debug Test — FAILED"))
		t.AppendRows([]table.Row{
			{"Host", fmt.Sprintf("%s:%d", sshHost, sshPort)},
			{"User", sshUser},
			{"Duration", elapsed.Round(time.Millisecond)},
			{"Error", err.Error()},
		})
		fmt.Println(t.Render())
		fmt.Println()
		for _, line := range logLines {
			fmt.Println(line)
		}
		return err
	}

	t.SetTitle(text.FgGreen.Sprint("SSH Debug Test — SUCCESS"))
	t.AppendRows([]table.Row{
		{"Host", fmt.Sprintf("%s:%d", sshHost, sshPort)},
		{"User", sshUser},
		{"Duration", elapsed.Round(time.Millisecond)},
		{"Exit Code", result.ExitCode},
		{"Output", strings.TrimSpace(result.Stdout)},
	})
	if result.Stderr != "" {
		t.AppendRow(table.Row{"Stderr", strings.TrimSpace(result.Stderr)})
	}
	fmt.Println(t.Render())
	fmt.Println()
	for _, line := range logLines {
		fmt.Println(line)
	}
	return nil
}
