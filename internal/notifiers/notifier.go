package notifiers

import (
	"errors"

	"github.com/hibare/GoPG2S3Dump/internal/config"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNotifiersDisabled = errors.New("notifiers are disabled")
	ErrNotifierDisabled  = errors.New("notifier is disabled")
)

func runPreChecks() error {
	if !config.Current.Notifiers.Enabled {
		return ErrNotifiersDisabled
	}

	return nil
}

func NotifyBackupSuccess(databases int, key string) {
	if err := runPreChecks(); err != nil {
		log.Error(err)
		return
	}

	discordNotifyBackupSuccess(databases, key)
}

func NotifyBackupFailure(err error) {
	if err := runPreChecks(); err != nil {
		log.Error(err)
		return
	}

	discordNotifyBackupFailure(err)
}

func NotifyBackupDeleteFailure(err error) {
	if err := runPreChecks(); err != nil {
		log.Error(err)
		return
	}

	discordNotifyBackupDeleteFailure(err)
}
