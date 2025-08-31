//nolint:gci // ignore import grouping
package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/spf13/cobra"

	commonLogger "github.com/hibare/GoCommon/v2/pkg/logger"
	"github.com/hibare/stashly/internal/config"
)

// cfgFile holds the path to the config file.
var cfgFile string

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "stashly",
	Short: "Automated PostgreSQL backup tool with cloud storage support",
	Long: `Stashly is a simple yet powerful CLI tool that automates PostgreSQL backups.
It supports scheduling backups using cron expressions and storing them securely
on multiple backends such as Amazon S3, Google Drive, and other cloud storage providers.

With Stashly, you can:
  - Schedule recurring PostgreSQL backups with flexible cron syntax
  - Automatically upload dumps to cloud storage backends
  - Get notified of backup failures through integrated notifiers
  - Run in the background as a long-lived process.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// start cron job that runs Dump according to config.
		// cron runs in background; block forever.
		ctx := cmd.Context()

		// Load config.
		cfg, err := config.LoadConfig(ctx, cfgFile)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to load config", "error", err)
			os.Exit(1)
		}

		slog.InfoContext(ctx, "Starting scheduled backup", "cron", cfg.Backup.Cron)
		scheduler := gocron.NewScheduler(time.UTC)
		_, err = scheduler.Cron(cfg.Backup.Cron).Do(func() {
			if bErr := doBackup(ctx, cfg); bErr != nil {
				slog.ErrorContext(ctx, "Scheduled backup failed", "error", bErr)
			} else {
				slog.InfoContext(ctx, "Scheduled backup completed successfully")
			}
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to schedule backup", "error", err)
		}
		scheduler.StartBlocking()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	ctx := context.Background()
	rootCmd.SetContext(ctx)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/stashly/config.yaml)")
	cobra.OnInitialize(commonLogger.InitDefaultLogger)
}
