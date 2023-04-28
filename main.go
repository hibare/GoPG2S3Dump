package main

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-co-op/gocron"

	"github.com/hibare/GoPG2S3Dump/internal/config"
	"github.com/hibare/GoPG2S3Dump/internal/notifiers"
	"github.com/hibare/GoPG2S3Dump/internal/pgdump"
)

func main() {
	config.LoadConfig()

	s := gocron.NewScheduler(time.UTC)

	// Schedule backup job
	if _, err := s.Cron(config.Current.Backup.Cron).Do(func() {
		databases, key, err := pgdump.Dump()

		if err != nil {
			log.Errorf("Error backingup databases %s", err)
			notifiers.BackupFailedNotification(err.Error())
			return
		}
		notifiers.BackupSuccessfulNotification(databases, key)

		err = pgdump.PurgeDumps()
		if err != nil {
			log.Errorf("Error purging dumps %s", err)
			notifiers.BackupDeletionFailureNotification(err.Error())
			return
		}
	}); err != nil {
		log.Fatalf("Error cron: %v", err)
	}
	log.Infof("Scheduled backup job to run every %s", config.Current.Backup.Cron)

	s.StartBlocking()
}
