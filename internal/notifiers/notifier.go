package notifiers

import (
	"github.com/hibare/GoPG2S3Dump/internal/config"
	log "github.com/sirupsen/logrus"
)

func BackupSuccessfulNotification(databases int, key string) {

	if !config.Current.Notifiers.Enabled {
		log.Warn("Notifiers are disabled")
		return
	}

	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return
	} else if config.Current.Notifiers.Discord.Enabled {
		if err := DiscordBackupSuccessfulNotification(config.Current.Notifiers.Discord.Webhook, config.Current.Backup.Hostname, databases, key); err != nil {
			log.Errorf("Error sending Discord notification: %v", err)
		}
	}

}

func BackupFailedNotification(err string) {

	if !config.Current.Notifiers.Enabled {
		log.Warn("Notifiers are disabled")
		return
	}

	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return
	} else if config.Current.Notifiers.Discord.Enabled {
		if err := DiscordBackupFailedNotification(config.Current.Notifiers.Discord.Webhook, config.Current.Backup.Hostname, err); err != nil {
			log.Errorf("Error sending Discord notification: %v", err)
		}
	}

}

func BackupDeletionFailureNotification(err string) {

	if !config.Current.Notifiers.Enabled {
		log.Warn("Notifiers are disabled")
		return
	}

	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return
	} else if config.Current.Notifiers.Discord.Enabled {
		if err := DiscordBackupDeletionFailureNotification(config.Current.Notifiers.Discord.Webhook, config.Current.Backup.Hostname, err); err != nil {
			log.Errorf("Error sending Discord notification: %v", err)
		}
	}
}
