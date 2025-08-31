package cmd

import (
	"log/slog"
	"os"

	"github.com/hibare/stashly/internal/config"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Trigger a backup run immediately",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		// Load config
		cfg, err := config.LoadConfig(ctx, cfgFile)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to load config", "error", err)
			os.Exit(1)
		}

		slog.InfoContext(ctx, "Starting immediate backup")
		if bErr := doBackup(ctx, cfg); bErr != nil {
			slog.ErrorContext(ctx, "Backup failed", "error", bErr)
			return
		}
		slog.InfoContext(ctx, "Backup completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
