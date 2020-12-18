package commands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ContainerSolutions/jeeves/pkg/config"
	"github.com/ContainerSolutions/jeeves/pkg/kubernetes"
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
	repoType, repo, candidateId, err := parseArgs(event.Text)
	if checkErr(err) {
		_, _, _ = cfg.SlackApi.PostMessage(
			event.ChannelID,
			slack.MsgOptionText(fmt.Sprintf("%v err: %v", errMessage, err.Error()), false),
		)
		return err
	}
	err = kubernetes.CreateAnonymizastionJob(cfg, repoType, repo, candidateId)
	if checkErr(err) {
		_, _, _ = cfg.SlackApi.PostMessage(
			event.ChannelID,
			slack.MsgOptionText(fmt.Sprintf("%v err: %v", errMessage, err.Error()), false),
		)
		return err
	}
	_, _, _ = cfg.SlackApi.PostMessage(
		event.ChannelID,
		slack.MsgOptionText("API Anonymized", false),
	)
	return nil
}

// getLinkAndId Adhoc function that parses command args
func parseArgs(message string) (string, string, string, error) {
	repoType := checkRepo(message)
	res := strings.Replace(message, "gitlab.com/", "", -1)
	res = strings.Replace(res, "github.com/", "", -1)
	res = strings.Replace(res, "\u00a0", " ", -1)
	res = strings.Replace(res, "https://", "", -1)
	res = strings.Replace(res, "http://", "", -1)
	res = strings.Replace(res, "  ", " ", -1)
	res = strings.Replace(res, "<", "", -1)
	res = strings.Replace(res, ">", "", -1)
	args := strings.Split(res, " ")
	if len(args) != 2 {
		return "", "", "", fmt.Errorf("Wrong Format")
	}
	return repoType, strings.ToLower(args[0]), args[1], nil
}

func checkRepo(message string) string {
	if strings.Contains(message, "github") {
		return "github"
	}
	return "gitlab"
}

func checkErr(err error) bool {
	return err != nil
}
