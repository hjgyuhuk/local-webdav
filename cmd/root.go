package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var logger *slog.Logger

var rootCmd = &cobra.Command{
	Use:   "local-webdav",
	Short: "A WebDAV server that serves authorized local folders",
	Long:  "A WebDAV server that maps virtual paths to local directories without exposing absolute paths.",
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
