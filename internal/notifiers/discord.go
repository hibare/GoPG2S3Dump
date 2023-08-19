package notifiers

import (
	"fmt"
	"strconv"

	"github.com/hibare/GoCommon/v2/pkg/notifiers/discord"
	"github.com/hibare/GoPG2S3Dump/internal/config"
	"github.com/hibare/GoPG2S3Dump/internal/constants"
	log "github.com/sirupsen/logrus"
)

func runDiscordPreChecks() error {
	if !config.Current.Notifiers.Discord.Enabled {
		return ErrNotifierDisabled
	}
	return nil
}

func discordNotifyBackupSuccess(databases int, key string) {
	if err := runDiscordPreChecks(); err != nil {
		log.Error(err)
		return
	}

	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Color: 1498748,
				Fields: []discord.EmbedField{
					{
						Name:   "Key",
						Value:  key,
						Inline: false,
					},
					{
						Name:   "Databases",
						Value:  strconv.Itoa(databases),
						Inline: false,
					},
				},
			},
		},
		Components: []discord.Component{},
		Username:   constants.ProgramIdentifier,
		Content:    fmt.Sprintf("**PGDB Backup Successful** - *%s*", config.Current.Backup.Hostname),
	}

	if err := message.Send(config.Current.Notifiers.Discord.Webhook); err != nil {
		log.Error(err)
	}
}

func discordNotifyBackupFailure(err error) {
	if err := runDiscordPreChecks(); err != nil {
		log.Error(err)
		return
	}

	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Error",
				Description: err.Error(),
				Color:       14554702,
			},
		},
		Components: []discord.Component{},
		Username:   constants.ProgramIdentifier,
		Content:    fmt.Sprintf("**PGDB Backup Failed** - *%s*", config.Current.Backup.Hostname),
	}

	if err := message.Send(config.Current.Notifiers.Discord.Webhook); err != nil {
		log.Error(err)
	}
}

func discordNotifyBackupDeleteFailure(err error) {
	if err := runDiscordPreChecks(); err != nil {
		log.Error(err)
		return
	}

	message := discord.Message{
		Embeds: []discord.Embed{
			{
				Title:       "Error",
				Description: err.Error(),
				Color:       14590998,
			},
		},
		Components: []discord.Component{},
		Username:   constants.ProgramIdentifier,
		Content:    fmt.Sprintf("**PGDB Backup Deletion Failed** - *%s*", config.Current.Backup.Hostname),
	}

	if err := message.Send(config.Current.Notifiers.Discord.Webhook); err != nil {
		log.Error(err)
	}
}
