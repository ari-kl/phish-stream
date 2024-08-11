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
						case "classify-domain":

							data := strings.Split(action.SelectedOption.Value, ":")

							domain := data[0]
							classification := data[1]

							// TODO: Send to takedown services

							ClassifyMessage(callback.Channel.ID, callback.Container.MessageTs, domain, classification)
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

	_, _, _, err = client.SendMessage(
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
				slack.NewOptionsSelectBlockElement("static_select", slack.NewTextBlockObject("plain_text", "Select an item", true, false), "classify-domain",
					slack.NewOptionBlockObject(GenerateClassificationText(domain, "postal"), slack.NewTextBlockObject("plain_text", "Postal", true, false), nil),
					slack.NewOptionBlockObject(GenerateClassificationText(domain, "banking"), slack.NewTextBlockObject("plain_text", "Banking", true, false), nil),
					slack.NewOptionBlockObject(GenerateClassificationText(domain, "item_scams"), slack.NewTextBlockObject("plain_text", "Item Scams", true, false), nil),
					slack.NewOptionBlockObject(GenerateClassificationText(domain, "other"), slack.NewTextBlockObject("plain_text", "Other", true, false), nil),
				),
				slack.NewButtonBlockElement("dismiss-domain", domain, slack.NewTextBlockObject("plain_text", "Dismiss", false, false)),
			),
		),
	)

	if err != nil {
		util.Logger.Error("Failed to send message to Slack", "err", err.Error())
	}
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

func ClassifyMessage(channelID string, timestamp string, domain string, classification string) {
	client.SendMessage(channelID,
		slack.MsgOptionUpdate(timestamp),
		slack.MsgOptionBlocks(
			slack.NewRichTextBlock("",
				slack.NewRichTextSection(
					slack.NewRichTextSectionTextElement(
						fmt.Sprintf("%s - confirmed (%s)", domain, classification),
						&slack.RichTextSectionTextStyle{Italic: true},
					),
				),
			),
		),
	)
}

func GenerateClassificationText(domain string, classification string) string {
	return fmt.Sprintf("%s:%s", domain, classification)
}
