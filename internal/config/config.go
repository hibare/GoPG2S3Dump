package config

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/hibare/GoPG2S3Dump/internal/constants"
	"github.com/hibare/GoPG2S3Dump/internal/utils"
)

type PostgresConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     string `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint" mapstructure:"endpoint"`
	Region    string `yaml:"region" mapstructure:"region"`
	AccessKey string `yaml:"access-key" mapstructure:"access-key"`
	SecretKey string `yaml:"secret-key" mapstructure:"secret-key"`
	Bucket    string `yaml:"bucket" mapstructure:"bucket"`
	Prefix    string `yaml:"prefix" mapstructure:"prefix"`
}

type BackupConfig struct {
	Hostname       string `yaml:"-"`
	RetentionCount int    `yaml:"retention-count" mapstructure:"retention-count"`
	DateTimeLayout string `yaml:"date-time-layout" mapstructure:"date-time-layout"`
	Cron           string `yaml:"cron" mapstructure:"cron"`
	Encrypt        bool   `yaml:"encrypt" mapstructure:"encrypt"`
}

type GPGConfig struct {
	KeyServer string `yaml:"key-server" mapstructure:"key-server"`
	KeyID     string `yaml:"key-id" mapstructure:"key-id"`
}

type Encryption struct {
	GPG GPGConfig
}

type DiscordNotifierConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Webhook string `yaml:"webhook" mapstructure:"webhook"`
}

type NotifiersConfig struct {
	Enabled bool                  `yaml:"enabled" mapstructure:"enabled"`
	Discord DiscordNotifierConfig `yaml:"discord" mapstructure:"discord"`
}

type Config struct {
	Postgres   PostgresConfig  `yaml:"postgres" mapstructure:"postgres"`
	S3         S3Config        `yaml:"s3" mapstructure:"s3"`
	Backup     BackupConfig    `yaml:"backup" mapstructure:"backup"`
	Encryption Encryption      `yaml:"encryption" mapstructure:"encryption"`
	Notifiers  NotifiersConfig `yaml:"notifiers" mapstructure:"notifiers"`
}

var Current *Config

func LoadConfig() {
	preCheckConfigPath()

	viper.AddConfigPath(constants.ConfigDir)
	viper.SetConfigName(constants.ConfigFilename)
	viper.SetConfigType(constants.ConfigFileExtension)

	// Load the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Unmarshal the configuration file into a struct
	if err := viper.Unmarshal(&Current); err != nil {
		log.Fatalf("Error parsing YAML data: %v", err)
	}

	// Set default DateTimeLayout if missing
	if Current.Backup.DateTimeLayout == "" {
		log.Warnf("DateTimeLayout is not set, using default: %s", constants.DefaultDateTimeLayout)
		Current.Backup.DateTimeLayout = constants.DefaultDateTimeLayout
	}

	// Set RetentionCount if missing
	if Current.Backup.RetentionCount == 0 {
		log.Warnf("RetentionCount is not set, using default: %d", constants.DefaultRetentionCount)
		Current.Backup.RetentionCount = constants.DefaultRetentionCount
	}

	// Set Schedule if missing
	if Current.Backup.Cron == "" {
		log.Warnf("Schedule is not set, using default: %s", constants.DefaultCron)
		Current.Backup.Cron = constants.DefaultCron
	}

	// Check if encryption is enabled & encryption config is enabled
	if Current.Backup.Encrypt {
		if Current.Encryption.GPG.KeyServer == "" || Current.Encryption.GPG.KeyID == "" {
			log.Fatalf("Error backup encryption is enabled but encryption config is not set")
		}
	}

	Current.Backup.Hostname = utils.GetHostname()
}

func GetConfigFilePath(configRootDir string) string {
	return filepath.Join(configRootDir, fmt.Sprintf("%s.%s", constants.ConfigFilename, constants.ConfigFileExtension))
}

func preCheckConfigPath() {
	configRootDir := constants.ConfigDir
	configPath := GetConfigFilePath(configRootDir)

	if info, err := os.Stat(configRootDir); os.IsNotExist(err) {
		log.Warnf("Config directory does not exist, creating: %s", configRootDir)
		if err := os.MkdirAll(configRootDir, 0755); err != nil {
			log.Fatalf("Error creating config directory: %v", err)
			return
		}
	} else if !info.IsDir() {
		log.Fatalf("Config directory is not a directory: %s", configRootDir)
		return
	}

	if info, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Warnf("Config file does not exist, creating: %s", configPath)
		file, err := os.Create(configPath)
		if err != nil {
			log.Fatalf("Error creating config file: %v", err)
			return
		}
		defer file.Close()

		// Marshal empty config
		yamlBytes, err := yaml.Marshal(Config{})
		if err != nil {
			log.Fatalf("Error marshaling config: %v", err)
		}

		// Write the YAML output to a file
		if _, err := file.Write(yamlBytes); err != nil {
			log.Fatalf("Error writing config file: %v", err)
		}

	} else if info.IsDir() {
		log.Fatalf("Expected file, found directory: %s", configPath)
		return
	}
}
