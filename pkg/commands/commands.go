package commands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ContainerSolutions/jeeves/pkg/config"
	"github.com/ContainerSolutions/jeeves/pkg/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

const errMessage = "There was an Error Anonymizing Repo Please Contact Brendan"

func AnonymizeHandler(
	cfg *config.JeevesConfig,
	r *http.Request,
) error {
	event, err := slack.SlashCommandParse(r)
	if checkErr(err) {
		return err
	}
	url, candidateId, err := getLinkAndId(event.Text)
	err = kubernetes.CreateAnonymizastionJob(cfg, url, candidateId)
	if checkErr(err) {
		_, _, _ = cfg.SlackApi.PostMessage(
			event.ChannelID,
			slack.MsgOptionText(errMessage, false),
		)
		return err
	}
	_, _, _ = cfg.SlackApi.PostMessage(
		event.ChannelID,
		slack.MsgOptionText("API Anonymized", false),
	)
	return nil
}

func getLinkAndId(Message string) (string, string, error) {
	log.Printf("%v", Message)
	res := strings.Replace(Message, "https://gitlab.com/", "", -1)
	res = strings.Replace(res, "\u00a0", " ", -1)
	args := strings.Split(res, " ")
	log.Printf("%v", args)
	if len(args) != 2 {
		return "", "", fmt.Errorf("Wrong Format")
	}
	return args[0], args[1], nil
}

func checkErr(err error) bool {
	if err != nil {
		return true
	}
	return false
}
