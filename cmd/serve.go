package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/local-webdav/internal/config"
	"github.com/user/local-webdav/internal/server"
)

var configPath string

func init() {
	defaultPath, _ := config.DefaultPath()
	serveCmd.Flags().StringVarP(&configPath, "config", "c", defaultPath, "path to config file")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the WebDAV server",
	RunE: func(cmd *cobra.Command, args []string) error {
		created, err := config.EnsureConfig(configPath)
		if err != nil {
			logger.Error("failed to ensure config", "err", err)
			return err
		}
		if created {
			fmt.Fprintf(os.Stderr, "配置文件已生成: %s\n请编辑配置文件添加 [[shares]] 后重新运行。\n", configPath)
			return nil
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			logger.Error("failed to load config", "err", err)
			return err
		}

		srv := server.New(cfg, logger)

		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start()
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-errCh:
			return err
		case sig := <-sigCh:
			logger.Info("received signal", "signal", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		}
	},
}
