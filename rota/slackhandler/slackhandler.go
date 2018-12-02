package slackhandler

import (
	"github.com/nlopes/slack"
	"github.com/sky-uk/support-bot/rota"
)

func MemberOptions(teamHistory rota.TeamSupportHistory) slack.Msg {

	slackAttachmentOptions := make([]slack.AttachmentActionOption, 0)

	for _, history := range rota.OrderedList(teamHistory) {
		slackAttachmentOptions = append(slackAttachmentOptions, slack.AttachmentActionOption{
			Text:  history.Name + " - " + string(history.DaysSupported) + " supported days",
			Value: history.Name,
		})
	}

	var attachment = slack.Attachment{
		Text:       "Next person on support",
		Color:      "#f9a41b",
		CallbackID: "orderedRota",
		Actions: []slack.AttachmentAction{
			{
				Name:    "actionSelect",
				Type:    "select",
				Options: slackAttachmentOptions,
			},
		},
	}

	message := slack.Msg{
		Text:        "This is an ordered list of support rota in the ascending order of support days",
		Attachments: []slack.Attachment{attachment},
	}

	return message
}

//func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	...
//	action := message.Actions[0]
//	switch action.Name {
//	case actionSelect:
//		value := action.SelectedOptions[0].Value
//		// Overwrite original drop down message.
//		originalMessage := message.OriginalMessage
//		originalMessage.Attachments[0].Text = fmt.Sprintf(“OK
//		to
//		order % s ?”, strings.Title(value))
//		originalMessage.Attachments[0].Actions = []slack.AttachmentAction{
//			{
//				Name: actionStart,
//				Text: “Yes”,
//				Type: “button”,
//				Value: “start”,
//				Style: “primary”,
//			},
//			{
//				Name: actionCancel,
//				Text: “No”,
//				Type: “button”,
//				Style: “danger”,
//			},
//		}
//		w.Header().Add(“Content -
//		type”, “application / json”)
//		w.WriteHeader(http.StatusOK)
//		json.NewEncoder(w).Encode(&originalMessage)
//		return
//	case actionStart:
//		title := “:
//	ok:
//		your
//		order
//		was
//		submitted
//		! yay
//		!”
//		responseMessage(w, message.OriginalMessage, title, “”)
//		return
//	case actionCancel:
//		title := fmt.Sprintf(“:
//	x: @%s
//		canceled
//		the
//		request”, message.User.Name)
//		responseMessage(w, message.OriginalMessage, title, “”)
//		return
//	default:
//		log.Printf(“[ERROR] ]Invalid action was submitted: %s”, action.Name)
//	w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//}
