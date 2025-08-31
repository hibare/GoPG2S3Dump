package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLoadConfig_Defaults(t *testing.T) {
	ctx := t.Context()
	cfg, err := LoadConfig(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadConfig_WithEnvVars(t *testing.T) {
	// Set environment variables
	t.Setenv("STASHLY_POSTGRES_HOST", "env-host")
	t.Setenv("STASHLY_POSTGRES_USER", "env-user")
	t.Setenv("STASHLY_POSTGRES_PASSWORD", "env-pass")

	t.Logf("Env STASHLY_POSTGRES_HOST: %s", os.Getenv("STASHLY_POSTGRES_HOST"))
	ctx := t.Context()
	cfg, err := LoadConfig(ctx, "")
	t.Logf("Loaded host: %s", cfg.Postgres.Host)
	require.NoError(t, err)
	assert.Equal(t, "env-host", cfg.Postgres.Host)
	assert.Equal(t, "env-user", cfg.Postgres.User)
	assert.Equal(t, "env-pass", cfg.Postgres.Password)
}

func TestLoadConfig_WithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"postgres": map[string]string{
			"host": "file-host",
			"user": "file-user",
		},
		"backup": map[string]interface{}{
			"retention-count": 5,
			"encrypt":         true,
		},
		"encryption": map[string]interface{}{
			"gpg": map[string]string{
				"key-server": "hkp://keys.example.com",
				"key-id":     "123ABC",
			},
		},
		"notifiers": map[string]interface{}{
			"discord": map[string]interface{}{
				"enabled": true,
				"webhook": "https://discord.com/api/webhooks/test",
			},
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	assert.Equal(t, "file-host", cfg.Postgres.Host)
	assert.Equal(t, 5, cfg.Backup.RetentionCount)
	assert.True(t, cfg.Backup.Encrypt)
	assert.Equal(t, "hkp://keys.example.com", cfg.Encryption.GPG.KeyServer)
	assert.True(t, cfg.Notifiers.Discord.Enabled)
}

func TestLoadConfig_EncryptSanityCheck(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Encryption enabled but no key-server/key-id
	content := map[string]interface{}{
		"backup": map[string]interface{}{
			"encrypt": true,
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	_ = yaml.NewEncoder(f).Encode(content)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	// Should have been reset to false
	assert.False(t, cfg.Backup.Encrypt)
}

func TestLoadConfig_DiscordSanityCheck(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Discord enabled but no webhook
	content := map[string]interface{}{
		"notifiers": map[string]interface{}{
			"discord": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	_ = yaml.NewEncoder(f).Encode(content)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	// Should have been reset to false
	assert.False(t, cfg.Notifiers.Discord.Enabled)
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML
	err := os.WriteFile(configFile, []byte(":::notyaml"), 0o600)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_WithS3Config(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"s3": map[string]string{
			"endpoint":   "https://s3.example.com",
			"region":     "us-west-2",
			"access-key": "test-access-key",
			"secret-key": "test-secret-key",
			"bucket":     "test-bucket",
			"prefix":     "backups/",
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	assert.Equal(t, "https://s3.example.com", cfg.S3.Endpoint)
	assert.Equal(t, "us-west-2", cfg.S3.Region)
	assert.Equal(t, "test-access-key", cfg.S3.AccessKey)
	assert.Equal(t, "test-secret-key", cfg.S3.SecretKey)
	assert.Equal(t, "test-bucket", cfg.S3.Bucket)
	assert.Equal(t, "backups/", cfg.S3.Prefix)
}

func TestLoadConfig_WithLoggerConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"logger": map[string]string{
			"level": "debug",
			"mode":  "json",
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	assert.Equal(t, "debug", cfg.Logger.Level)
	assert.Equal(t, "json", cfg.Logger.Mode)
}

func TestLoadConfig_WithAppConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"app": map[string]string{
			"instance-id": "test-instance-123",
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	assert.Equal(t, "test-instance-123", cfg.App.InstanceID)
}

func TestLoadConfig_WithAllConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"postgres": map[string]string{
			"host":     "localhost",
			"port":     "5433",
			"user":     "testuser",
			"password": "testpass",
		},
		"s3": map[string]string{
			"endpoint":   "https://s3.example.com",
			"region":     "us-east-1",
			"access-key": "access123",
			"secret-key": "secret123",
			"bucket":     "mybucket",
			"prefix":     "prefix/",
		},
		"backup": map[string]interface{}{
			"retention-count":  10,
			"date-time-layout": "2006-01-02",
			"cron":             "0 2 * * *",
			"encrypt":          true,
		},
		"encryption": map[string]interface{}{
			"gpg": map[string]string{
				"key-server": "hkp://keyserver.ubuntu.com",
				"key-id":     "ABC123",
			},
		},
		"notifiers": map[string]interface{}{
			"enabled": true,
			"discord": map[string]interface{}{
				"enabled": true,
				"webhook": "https://discord.com/api/webhooks/test",
			},
		},
		"logger": map[string]string{
			"level": "info",
			"mode":  "text",
		},
		"app": map[string]string{
			"instance-id": "full-config-test",
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	// Verify all fields are set correctly
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, "5433", cfg.Postgres.Port)
	assert.Equal(t, "testuser", cfg.Postgres.User)
	assert.Equal(t, "testpass", cfg.Postgres.Password)

	assert.Equal(t, "https://s3.example.com", cfg.S3.Endpoint)
	assert.Equal(t, "us-east-1", cfg.S3.Region)
	assert.Equal(t, "access123", cfg.S3.AccessKey)
	assert.Equal(t, "secret123", cfg.S3.SecretKey)
	assert.Equal(t, "mybucket", cfg.S3.Bucket)
	assert.Equal(t, "prefix/", cfg.S3.Prefix)

	assert.Equal(t, 10, cfg.Backup.RetentionCount)
	assert.Equal(t, "2006-01-02", cfg.Backup.DateTimeLayout)
	assert.Equal(t, "0 2 * * *", cfg.Backup.Cron)
	assert.True(t, cfg.Backup.Encrypt)

	assert.Equal(t, "hkp://keyserver.ubuntu.com", cfg.Encryption.GPG.KeyServer)
	assert.Equal(t, "ABC123", cfg.Encryption.GPG.KeyID)

	assert.True(t, cfg.Notifiers.Enabled)
	assert.True(t, cfg.Notifiers.Discord.Enabled)
	assert.Equal(t, "https://discord.com/api/webhooks/test", cfg.Notifiers.Discord.Webhook)

	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, "text", cfg.Logger.Mode)

	assert.Equal(t, "full-config-test", cfg.App.InstanceID)
}

func TestLoadConfig_EnvironmentVariableOverrides(t *testing.T) {
	// Set environment variables that should override config file
	t.Setenv("STASHLY_POSTGRES_HOST", "env-override-host")
	t.Setenv("STASHLY_S3_BUCKET", "env-override-bucket")

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"postgres": map[string]string{
			"host": "file-host",
		},
		"s3": map[string]string{
			"bucket": "file-bucket",
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	// Environment variables should take precedence
	assert.Equal(t, "env-override-host", cfg.Postgres.Host)
	assert.Equal(t, "env-override-bucket", cfg.S3.Bucket)
}

func TestLoadConfig_EmptyConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create empty config file
	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadConfig_ConfigFileNotFound(t *testing.T) {
	ctx := t.Context()
	cfg, err := LoadConfig(ctx, "/nonexistent/config.yaml")
	require.Error(t, err) // Should error when specific config file doesn't exist
	assert.Nil(t, cfg)
}

func TestLoadConfig_WithContext(t *testing.T) {
	ctx := context.Background()
	cfg, err := LoadConfig(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadConfig_EncryptionSanityCheckWithValidKeys(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"backup": map[string]interface{}{
			"encrypt": true,
		},
		"encryption": map[string]interface{}{
			"gpg": map[string]string{
				"key-server": "hkp://keyserver.ubuntu.com",
				"key-id":     "VALID123",
			},
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	// Encryption should remain enabled
	assert.True(t, cfg.Backup.Encrypt)
	assert.Equal(t, "hkp://keyserver.ubuntu.com", cfg.Encryption.GPG.KeyServer)
	assert.Equal(t, "VALID123", cfg.Encryption.GPG.KeyID)
}

func TestLoadConfig_DiscordSanityCheckWithValidWebhook(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	content := map[string]interface{}{
		"notifiers": map[string]interface{}{
			"enabled": true,
			"discord": map[string]interface{}{
				"enabled": true,
				"webhook": "https://discord.com/api/webhooks/valid123",
			},
		},
	}

	//nolint:gosec // Safe in tests - using t.TempDir()
	f, err := os.Create(configFile)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	err = yaml.NewEncoder(f).Encode(content)
	require.NoError(t, err)

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, configFile)
	require.NoError(t, err)

	// Discord should remain enabled
	assert.True(t, cfg.Notifiers.Enabled)
	assert.True(t, cfg.Notifiers.Discord.Enabled)
	assert.Equal(t, "https://discord.com/api/webhooks/valid123", cfg.Notifiers.Discord.Webhook)
}

func TestLoadConfig_EnvironmentVariablePriority(t *testing.T) {
	// Test that environment variables have higher priority than defaults
	t.Setenv("STASHLY_POSTGRES_PORT", "5434")
	t.Setenv("STASHLY_BACKUP_RETENTION_COUNT", "15")

	ctx := t.Context()
	cfg, err := LoadConfig(ctx, "")
	require.NoError(t, err)

	// Environment variables should override defaults
	assert.Equal(t, "5434", cfg.Postgres.Port)
	assert.Equal(t, 15, cfg.Backup.RetentionCount)
}
