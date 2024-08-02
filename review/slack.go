package review

import (
	"fmt"
	"os"
	"strings"

	"github.com/ari-kl/phish-stream/shared"
	"github.com/ari-kl/phish-stream/util"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// We use these across multiple functions, so they need to be package-level
var api = slack.New(
	os.Getenv("SLACK_BOT_TOKEN"),
	slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
)
var client = socketmode.New(api)

func StartSlackBot() {
	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				util.Logger.Info("Connecting to Slack")
			case socketmode.EventTypeConnected:
				util.Logger.Info("Connected to Slack")
			case socketmode.EventTypeConnectionError:
				util.Logger.Error("Error connecting to slack")
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)

				if !ok {
					util.Logger.Error("Failed to cast evt.Data to slack.InteractionCallback")
					continue
				}

				switch callback.Type {
				case slack.InteractionTypeBlockActions:
					for _, action := range callback.ActionCallback.BlockActions {
						switch action.ActionID {
						case "confirm-domain":
							// TODO: Send to takedown services
							ConfirmMessage(callback.Channel.ID, callback.Container.MessageTs, action.Value)
						case "dismiss-domain":
							// All we need to do is dismiss the message, no further action required for dismissed domains
							DismissMessage(callback.Channel.ID, callback.Container.MessageTs, action.Value)
						}
					}
				}

				client.Ack(*evt.Request)
			}
		}
	}()

	client.Run()
}

func SendMessage(domain string, result shared.FilterResult) {
	var matchText string

	switch result.MatchType {
	case shared.FilterMatchTypeKeyword:
		matchText = fmt.Sprintf("Keyword: \"%s\"", result.MatchedBy)
	case shared.FilterMatchTypeSimilarity:
		matchText = fmt.Sprintf("Term: \"%s\" (%.3f)", result.MatchedBy, result.SimilarityScore)
	case shared.FilterMatchTypeRegex:
		matchText = fmt.Sprintf("Pattern: \"/%s/\"", result.MatchedBy)
	}

	var hostText string
	err, isp, country := LookupISP(domain)

	if err != nil {
		hostText = "Host: Unknown"
	} else {
		hostText = fmt.Sprintf("Host: %s (:flag-%s:)", isp, strings.ToLower(country))
	}

	client.SendMessage(
		os.Getenv("SLACK_CHANNEL_ID"),
		slack.MsgOptionBlocks(
			slack.NewRichTextBlock("",
				slack.NewRichTextSection(
					slack.NewRichTextSectionTextElement(domain, &slack.RichTextSectionTextStyle{}),
				),
			),
			slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Filter: \"%s\"\t%s", result.Name, matchText), false, false)),
			slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", hostText, false, false)),
			slack.NewActionBlock("",
				slack.NewButtonBlockElement("confirm-domain", domain, slack.NewTextBlockObject("plain_text", "Confirm", false, false)).WithStyle(slack.StyleDanger),
				slack.NewButtonBlockElement("dismiss-domain", domain, slack.NewTextBlockObject("plain_text", "Dismiss", false, false)),
			),
		),
	)
}

func DismissMessage(channelID string, timestamp string, domain string) {
	client.SendMessage(channelID,
		slack.MsgOptionUpdate(timestamp),
		slack.MsgOptionBlocks(
			slack.NewRichTextBlock("",
				slack.NewRichTextSection(
					slack.NewRichTextSectionTextElement(
						fmt.Sprintf("%s - dismissed", domain),
						&slack.RichTextSectionTextStyle{Italic: true},
					),
				),
			),
		),
	)
}

func ConfirmMessage(channelID string, timestamp string, domain string) {
	client.SendMessage(channelID,
		slack.MsgOptionUpdate(timestamp),
		slack.MsgOptionBlocks(
			slack.NewRichTextBlock("",
				slack.NewRichTextSection(
					slack.NewRichTextSectionTextElement(
						fmt.Sprintf("%s - confirmed", domain),
						&slack.RichTextSectionTextStyle{Italic: true},
					),
				),
			),
		),
	)
}
