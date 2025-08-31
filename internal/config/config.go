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
	configEnvPrefix       = "STASHLY"
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

// Current holds the currently loaded configuration.
var Current *Config

// LoadConfig loads config from viper.
func LoadConfig(ctx context.Context, configPath string) (*Config, error) {
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

	// Add defaults
	v.SetDefault("backup.retention-count", constants.DefaultRetentionCount)
	v.SetDefault("backup.date-time-layout", constants.DefaultDateTimeLayout)
	v.SetDefault("backup.cron", constants.DefaultCron)
	v.SetDefault("logger.level", commonLogger.DefaultLoggerLevel)
	v.SetDefault("logger.mode", commonLogger.DefaultLoggerMode)
	v.SetDefault("app.instance-id", commonUtils.GetHostname())

	// Environment variable binding (STASHLY_POSTGRES_HOST, etc.)
	v.SetEnvPrefix(configEnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

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

	// Unmarshal into Current
	if err := v.Unmarshal(&Current); err != nil {
		return nil, err
	}

	// Initialize logger
	commonLogger.InitLogger(&Current.Logger.Level, &Current.Logger.Mode)

	// Encryption sanity check
	if Current.Backup.Encrypt {
		if Current.Encryption.GPG.KeyServer == "" || Current.Encryption.GPG.KeyID == "" {
			slog.WarnContext(ctx, "GPG encryption enabled but key-server/key-id not set; disabling encryption")
			Current.Backup.Encrypt = false
		}
	}

	// Notifiers sanity check
	if Current.Notifiers.Discord.Webhook == "" {
		slog.WarnContext(ctx, "Discord notifier disabled (missing webhook)")
		Current.Notifiers.Discord.Enabled = false
	}

	return Current, nil
}
