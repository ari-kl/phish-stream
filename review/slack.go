package review

import (
	"fmt"
	"os"

	"github.com/ari-kl/phish-stream/util"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// We use these outside of StartSlackBot, so they need to be package-level
var api = slack.New(os.Getenv("SLACK_BOT_TOKEN"), slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")))
var client = socketmode.New(api, socketmode.OptionDebug(true))

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
							util.Logger.Info("Confirmed domain: " + action.Value)
						case "dismiss-domain":
							util.Logger.Info("Dismissed domain: " + action.Value)
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

func SendMessage(domain string, filter string) {
	client.SendMessage(
		os.Getenv("SLACK_CHANNEL_ID"),
		slack.MsgOptionCompose(
			slack.MsgOptionText(fmt.Sprintf("Matched: \"%s\"", filter), false),
			slack.MsgOptionBlocks(
				slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Matched: \"%s\"", filter), false, false)),
				slack.NewActionBlock("",
					slack.NewButtonBlockElement("confirm-domain", domain, slack.NewTextBlockObject("plain_text", "Confirm", false, false)).WithStyle(slack.StyleDanger),
					slack.NewButtonBlockElement("dismiss-domain", domain, slack.NewTextBlockObject("plain_text", "Dismiss", false, false)),
				),
			),
		),
	)
}

func DismissMessage(channelID string, timestamp string, domain string) {
	client.SendMessage(channelID,
		slack.MsgOptionUpdate(timestamp),
		slack.MsgOptionText(fmt.Sprintf("Dismissed: \"%s\"", domain), false),
	)
}
