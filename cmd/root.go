package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	appVersion string
	buildTime  string
	gitCommit  string
)

var rootCmd = &cobra.Command{
	Use:   "gorundeck",
	Short: "Go-Rundeck – runbook automation platform",
	Long:  `Go-Rundeck is a web-based runbook automation and task orchestration platform built natively in Go.`,
}

// Execute is the entry point called from main.go.
func Execute(version, bt, commit string) {
	appVersion = version
	buildTime = bt
	gitCommit = commit

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.toml", "path to configuration file")
}
