<div align="center">
  <img src="internal/assets/favicon.png" alt="Stashly Logo" width="200"/>
  <h3>Secure, automated PostgreSQL backups for the modern cloud era</h3>
</div>

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.24.4+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://hub.docker.com/r/hibare/stashly)
[![Go Report Card](https://goreportcard.com/badge/github.com/hibare/stashly)](https://goreportcard.com/report/github.com/hibare/stashly)
[![Docker Hub](https://img.shields.io/docker/pulls/hibare/stashly)](https://hub.docker.com/r/hibare/stashly)
[![Docker image size](https://img.shields.io/docker/image-size/hibare/stashly/latest)](https://hub.docker.com/r/hibare/stashly)
[![GitHub issues](https://img.shields.io/github/issues/hibare/stashly)](https://github.com/hibare/stashly/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/hibare/stashly)](https://github.com/hibare/stashly/issues)
[![GitHub](https://img.shields.io/github/license/hibare/stashly)](https://github.com/hibare/stashly/blob/main/LICENSE)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/hibare/stashly)](https://github.com/hibare/stashly/releases)

</div>

---

**Stashly** is a powerful, automated PostgreSQL backup tool with cloud storage support. It provides scheduled backups, encryption, and seamless integration with S3-compatible storage backends.

## 🚀 Features

- **Automated PostgreSQL Backups**: Schedule recurring backups using cron expressions
- **Cloud Storage Integration**: Upload backups to S3-compatible storage (AWS S3, MinIO, etc.)
- **GPG Encryption**: Optional GPG encryption for enhanced security
- **Smart Retention Policy**: Automatically manage backup retention and cleanup
- **Discord Notifications**: Get notified of backup success/failure via Discord webhooks
- **Docker Support**: Ready-to-use Docker images for easy deployment
- **CLI Interface**: Simple command-line interface with immediate backup triggers
- **Multi-Database Support**: Automatically detect and backup all non-template databases

## 📋 Requirements

- **Go 1.24.4+** (for building from source)
- **PostgreSQL client tools** (`psql`, `pg_dump`)
- **S3-compatible storage** (AWS S3, MinIO, etc.)
- **GPG key server** (if using encryption)

## 🛠️ Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/hibare/stashly.git
cd stashly

# Build the binary
go build -o stashly main.go

# Install to system PATH (optional)
sudo cp stashly /usr/local/bin/
```

### Using Docker

```bash
# Pull the official image
docker pull hibare/stashly

# Or build locally
docker build -t stashly .
```

## ⚙️ Configuration

Stashly uses YAML configuration files. Create a config file at `/etc/stashly/config.yaml` or specify a custom path.

### Configuration File Structure

```yaml
# PostgreSQL connection settings
postgres:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "your_password"

# S3 storage configuration
s3:
  endpoint: "https://s3.amazonaws.com" # or your S3-compatible endpoint
  region: "us-east-1"
  access-key: "your_access_key"
  secret-key: "your_secret_key"
  bucket: "your_backup_bucket"
  prefix: "postgres_backups"

# Backup settings
backup:
  retention-count: 30 # Number of backups to retain
  cron: "0 0 * * *" # Cron schedule (daily at midnight)
  encrypt: false # Enable GPG encryption

# GPG encryption (if enabled)
encryption:
  gpg:
    key-server: "keyserver.ubuntu.com"
    key-id: "your_gpg_key_id"

# Notifications
notifiers:
  enabled: true
  discord:
    enabled: true
    webhook: "your_discord_webhook_url"

# Logging
logger:
  level: "info"
  mode: "json"
```

### Environment Variables

All configuration options can be set via environment variables using the `STASHLY_` prefix:

```bash
export STASHLY_POSTGRES_HOST=localhost
export STASHLY_POSTGRES_PORT=5432
export STASHLY_POSTGRES_USER=postgres
export STASHLY_POSTGRES_PASSWORD=your_password
export STASHLY_S3_ENDPOINT=https://s3.amazonaws.com
export STASHLY_S3_REGION=us-east-1
export STASHLY_S3_ACCESS_KEY=your_access_key
export STASHLY_S3_SECRET_KEY=your_secret_key
export STASHLY_S3_BUCKET=your_backup_bucket
export STASHLY_S3_PREFIX=postgres_backups
export STASHLY_BACKUP_CRON="0 0 * * *"
export STASHLY_BACKUP_RETENTION_COUNT=30
export STASHLY_BACKUP_ENCRYPT=false
export STASHLY_NOTIFIERS_DISCORD_WEBHOOK=your_discord_webhook_url
```

## 🚀 Usage

### Command Line Interface

Stashly provides a simple CLI with the following commands:

```bash
# Start scheduled backups (default behavior)
stashly

# Trigger an immediate backup
stashly backup

# Use custom config file
stashly --config /path/to/config.yaml

# Start with scheduled backups
stashly --config /path/to/config.yaml
```

### Docker Usage

```bash
# Run with custom config
docker run -v /path/to/config:/etc/stashly/config.yaml hibare/stashly

# Run with environment variables
docker run \
  -e STASHLY_POSTGRES_HOST=host.docker.internal \
  -e STASHLY_POSTGRES_USER=postgres \
  -e STASHLY_POSTGRES_PASSWORD=password \
  -e STASHLY_S3_ENDPOINT=https://s3.amazonaws.com \
  -e STASHLY_S3_REGION=us-east-1 \
  -e STASHLY_S3_ACCESS_KEY=your_key \
  -e STASHLY_S3_SECRET_KEY=your_secret \
  -e STASHLY_S3_BUCKET=backups \
  hibare/stashly
```

### Docker Compose

```yaml
version: "3.9"

services:
  stashly:
    image: hibare/stashly
    container_name: stashly
    volumes:
      - ./config:/etc/stashly
    environment:
      - STASHLY_POSTGRES_HOST=postgres
      - STASHLY_POSTGRES_USER=postgres
      - STASHLY_POSTGRES_PASSWORD=password
    networks:
      - db
    depends_on:
      - postgres

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:

networks:
  db:
    driver: bridge
```

## 🔧 Development

### Prerequisites

- Go 1.24.4+
- Docker and Docker Compose
- PostgreSQL client tools

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/hibare/stashly.git
cd stashly

# Initialize development environment
make init

# Start development services (PostgreSQL + MinIO)
make dev

# Run tests
make test

# Clean up
make clean
```

### Project Structure

```
stashly/
├── cmd/                    # Command-line interface
│   ├── backup.go          # Backup command implementation
│   ├── common.go          # Common functionality
│   └── root.go            # Root command and scheduling
├── internal/               # Internal packages
│   ├── assets/            # Application assets (logo, etc.)
│   ├── config/            # Configuration management
│   ├── constants/         # Application constants
│   ├── dumpster/          # PostgreSQL dump functionality
│   ├── exec/              # Command execution interface
│   ├── notifiers/         # Notification services
│   │   └── discord/       # Discord notification implementation
│   └── storage/           # Storage backends
│       └── s3/            # S3 storage implementation
├── testhelpers/           # Test utilities
├── docker-compose.yml     # Production Docker setup
├── docker-compose.dev.yml # Development environment
├── Dockerfile             # Multi-stage Docker build
├── Makefile               # Development tasks
└── main.go                # Application entry point
```

### Building

```bash
# Build for current platform
go build -o stashly main.go

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o stashly main.go

# Build with Docker
docker build -t stashly .
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/dumpster/...
```

## 📊 Backup Process

1. **Pre-flight Checks**: Verify PostgreSQL tools availability and create temporary directories
2. **Database Discovery**: Automatically detect all non-template databases
3. **Dump Creation**: Create SQL dumps using `pg_dump` for each database
4. **Archive Creation**: Compress all dumps into a single archive
5. **Encryption** (optional): Encrypt the archive using GPG if enabled
6. **Upload**: Upload to configured storage backend
7. **Cleanup**: Remove temporary files and old backups based on retention policy
8. **Notification**: Send success/failure notifications via configured notifiers

## 🔐 Security Features

- **GPG Encryption**: Optional GPG encryption for backup files
- **Secure Storage**: Support for S3-compatible storage with access controls
- **Environment Variables**: Secure configuration via environment variables
- **Temporary Files**: Automatic cleanup of temporary backup files

## 📈 Monitoring and Notifications

### Discord Notifications

Stashly can send notifications to Discord channels via webhooks:

- **Backup Success**: Database count and storage location
- **Backup Failure**: Error details and failure information
- **Cleanup Failure**: Retention policy cleanup errors

### Logging

Comprehensive logging with configurable levels:

- **JSON Mode**: Structured logging for production environments, `PRETTY` / `JSON`
- **Text Mode**: Human-readable logs for development
- **Configurable Levels**: Debug, Info, Warn, Error, `DEBUG` / `INFO` / `ERROR`

## 🐳 Docker Development Environment

The project includes a complete development environment with:

- **PostgreSQL 16**: Database server for testing
- **MinIO**: S3-compatible object storage
- **Pre-configured buckets**: Ready-to-use storage buckets
- **Network isolation**: Secure development environment

Start the development environment:

```bash
make dev
```

This will start PostgreSQL on port 5432 and MinIO on ports 9000 (API) and 9001 (Console).

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
