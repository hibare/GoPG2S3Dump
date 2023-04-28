package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hibare/GoPG2S3Dump/internal/config"
	log "github.com/sirupsen/logrus"
)

const DiscordUser = "PGDB Backup Job"

type DiscordWebhookMessage struct {
	Embeds     []DiscordEmbed     `json:"embeds"`
	Components []DiscordComponent `json:"components"`
	Username   string             `json:"username"`
	Content    string             `json:"content"`
}

type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Footer      DiscordEmbedFooter  `json:"footer"`
	Fields      []DiscordEmbedField `json:"fields"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

type DiscordComponent struct {
	// Define struct for Discord components if needed
}

func DiscordBackupSuccessfulNotification(webhookUrl string, hostname string, databases int, key string) error {
	webhookMessage := DiscordWebhookMessage{
		Embeds: []DiscordEmbed{
			{
				Color: 1498748,
				Fields: []DiscordEmbedField{
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
		Components: []DiscordComponent{},
		Username:   DiscordUser,
		Content:    fmt.Sprintf("**PGDB Backup Successful** - *%s*", hostname),
	}

	return SendMessage(webhookUrl, webhookMessage)

}

func DiscordBackupFailedNotification(webhookUrl string, hostname, err string) error {
	webhookMessage := DiscordWebhookMessage{
		Embeds: []DiscordEmbed{
			{
				Title:       "Error",
				Description: err,
				Color:       14554702,
			},
		},
		Components: []DiscordComponent{},
		Username:   DiscordUser,
		Content:    fmt.Sprintf("**PGDB Backup Failed** - *%s*", hostname),
	}

	return SendMessage(webhookUrl, webhookMessage)
}

func DiscordBackupDeletionFailureNotification(webhookUrl string, hostname, err string) error {
	webhookMessage := DiscordWebhookMessage{
		Embeds: []DiscordEmbed{
			{
				Title:       "Error",
				Description: err,
				Color:       14590998,
			},
		},
		Components: []DiscordComponent{},
		Username:   DiscordUser,
		Content:    fmt.Sprintf("**PGDB Backup Deletion Failed** - *%s*", hostname),
	}

	return SendMessage(webhookUrl, webhookMessage)
}

func SendMessage(webhookUrl string, message DiscordWebhookMessage) error {
	if config.Current.Notifiers.Discord.Webhook != "" && !config.Current.Notifiers.Discord.Enabled {
		log.Warning("Discord notifier not enabled")
		return nil
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return &json.SyntaxError{}
	}

	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
