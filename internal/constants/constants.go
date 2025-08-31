// Package constants defines application-wide constant values.
package constants

const (
	// ProgramIdentifier is the name used in notifications and logs.
	ProgramIdentifier = "Stashly"

	// ExportDir is the directory where database exports are temporarily stored.
	ExportDir = "db_exports"

	// DefaultDateTimeLayout is the default layout for datetime strings in backup filenames.
	DefaultDateTimeLayout = "20060102150405"

	// DefaultRetentionCount is the default number of backups to retain.
	DefaultRetentionCount = 30

	//  DefaultCron is the default cron schedule for backups (daily at midnight).
	DefaultCron = "0 0 * * *"
)
