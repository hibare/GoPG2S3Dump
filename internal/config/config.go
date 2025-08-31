// Package config handles application configuration loading and management.
package config

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	commonLogger "github.com/hibare/GoCommon/v2/pkg/logger"
	commonUtils "github.com/hibare/GoCommon/v2/pkg/utils"
	"github.com/hibare/stashly/internal/constants"
	"github.com/spf13/viper"
)

const (
	configFileName        = "config"
	configFileType        = "yaml"
	configFileDefaultPath = "/etc/stashly/"
)

// AppConfig holds application-level configuration.
type AppConfig struct {
	InstanceID string `mapstructure:"instance-id"`
}

// LoggerConfig holds logging configuration.
type LoggerConfig struct {
	Level string `mapstructure:"level"`
	Mode  string `mapstructure:"mode"`
}

// PostgresConfig holds PostgreSQL connection configuration.
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// S3Config holds S3 storage configuration.
type S3Config struct {
	Endpoint  string `mapstructure:"endpoint"`
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"access-key"`
	SecretKey string `mapstructure:"secret-key"`
	Bucket    string `mapstructure:"bucket"`
	Prefix    string `mapstructure:"prefix"`
}

// BackupConfig holds backup-related configuration.
type BackupConfig struct {
	RetentionCount int    `mapstructure:"retention-count"`
	DateTimeLayout string `mapstructure:"date-time-layout"`
	Cron           string `mapstructure:"cron"`
	Encrypt        bool   `mapstructure:"encrypt"`
}

// GPGConfig holds GPG encryption configuration.
type GPGConfig struct {
	KeyServer string `mapstructure:"key-server"`
	KeyID     string `mapstructure:"key-id"`
}

// Encryption holds encryption-related configuration.
type Encryption struct {
	GPG GPGConfig `mapstructure:"gpg"`
}

// DiscordNotifierConfig holds configuration for the Discord notifier.
type DiscordNotifierConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
}

// NotifiersConfig holds configuration for all notifiers.
type NotifiersConfig struct {
	Enabled bool                  `mapstructure:"enabled"`
	Discord DiscordNotifierConfig `mapstructure:"discord"`
}

// Config is the main configuration struct that holds all configuration sections.
type Config struct {
	App        AppConfig       `mapstructure:"app"`
	Postgres   PostgresConfig  `mapstructure:"postgres"`
	S3         S3Config        `mapstructure:"s3"`
	Backup     BackupConfig    `mapstructure:"backup"`
	Encryption Encryption      `mapstructure:"encryption"`
	Notifiers  NotifiersConfig `mapstructure:"notifiers"`
	Logger     LoggerConfig    `mapstructure:"logger"`
}

// LoadConfig loads config from viper.
func LoadConfig(ctx context.Context, configPath string) (*Config, error) {
	var cfg *Config
	v := viper.New()
	v.SetConfigName(configFileName)
	v.SetConfigType(configFileType)

	// Config search paths
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath(configFileDefaultPath)
	}

	// Environment variable binding (STASHLY_POSTGRES_HOST, etc.)
	v.SetEnvPrefix("STASHLY")
	v.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`, `-`, `_`))
	v.AutomaticEnv()

	// Bind all configuration fields to environment variables
	envBindings := map[string]string{
		"postgres.host":             "STASHLY_POSTGRES_HOST",
		"postgres.port":             "STASHLY_POSTGRES_PORT",
		"postgres.user":             "STASHLY_POSTGRES_USER",
		"postgres.password":         "STASHLY_POSTGRES_PASSWORD",
		"s3.endpoint":               "STASHLY_S3_ENDPOINT",
		"s3.region":                 "STASHLY_S3_REGION",
		"s3.access-key":             "STASHLY_S3_ACCESS_KEY",
		"s3.secret-key":             "STASHLY_S3_SECRET_KEY",
		"s3.bucket":                 "STASHLY_S3_BUCKET",
		"s3.prefix":                 "STASHLY_S3_PREFIX",
		"backup.retention-count":    "STASHLY_BACKUP_RETENTION_COUNT",
		"backup.date-time-layout":   "STASHLY_BACKUP_DATE_TIME_LAYOUT",
		"backup.cron":               "STASHLY_BACKUP_CRON",
		"backup.encrypt":            "STASHLY_BACKUP_ENCRYPT",
		"encryption.gpg.key-server": "STASHLY_ENCRYPTION_GPG_KEY_SERVER",
		"encryption.gpg.key-id":     "STASHLY_ENCRYPTION_GPG_KEY_ID",
		"notifiers.enabled":         "STASHLY_NOTIFIERS_ENABLED",
		"notifiers.discord.enabled": "STASHLY_NOTIFIERS_DISCORD_ENABLED",
		"notifiers.discord.webhook": "STASHLY_NOTIFIERS_DISCORD_WEBHOOK",
		"logger.level":              "STASHLY_LOGGER_LEVEL",
		"logger.mode":               "STASHLY_LOGGER_MODE",
		"app.instance-id":           "STASHLY_APP_INSTANCE_ID",
	}

	for configKey, envVar := range envBindings {
		if err := v.BindEnv(configKey, envVar); err != nil {
			slog.WarnContext(ctx, "Failed to bind environment variable",
				slog.String("config", configKey),
				slog.String("env", envVar),
				slog.String("error", err.Error()))
		}
	}

	// Try read config
	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundErr) {
			slog.WarnContext(ctx, "No config file found, relying on env vars/defaults")
		} else {
			return nil, err
		}
	} else {
		slog.InfoContext(ctx, "Using config file", slog.String("file", v.ConfigFileUsed()))
	}

	// Add defaults
	v.SetDefault("postgres.host", constants.DefaultPostgresHost)
	v.SetDefault("postgres.port", constants.DefaultPostgresPort)
	v.SetDefault("postgres.port", "5432")
	v.SetDefault("backup.retention-count", constants.DefaultRetentionCount)
	v.SetDefault("backup.date-time-layout", constants.DefaultDateTimeLayout)
	v.SetDefault("backup.cron", constants.DefaultCron)
	v.SetDefault("logger.level", commonLogger.DefaultLoggerLevel)
	v.SetDefault("logger.mode", commonLogger.DefaultLoggerMode)
	v.SetDefault("app.instance-id", commonUtils.GetHostname())

	// Unmarshal into Current
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Initialize logger
	commonLogger.InitLogger(&cfg.Logger.Level, &cfg.Logger.Mode)

	// Encryption sanity check
	if cfg.Backup.Encrypt {
		if cfg.Encryption.GPG.KeyServer == "" || cfg.Encryption.GPG.KeyID == "" {
			slog.WarnContext(ctx, "GPG encryption enabled but key-server/key-id not set; disabling encryption")
			cfg.Backup.Encrypt = false
		}
	}

	// Notifiers sanity check
	if cfg.Notifiers.Discord.Enabled {
		if cfg.Notifiers.Discord.Webhook == "" {
			slog.WarnContext(ctx, "Discord notifier enabled but missing webhook; disabling notifier")
			cfg.Notifiers.Discord.Enabled = false
		}
	}

	return cfg, nil
}
