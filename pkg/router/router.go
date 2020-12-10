package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ContainerSolutions/jeeves/pkg/commands"
	"github.com/ContainerSolutions/jeeves/pkg/config"
	"github.com/ContainerSolutions/jeeves/pkg/kubernetes"
	jslack "github.com/ContainerSolutions/jeeves/pkg/slack"
	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// GetRouter returns the routes for the server
func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/anonymize", anonEventHandler)
	r.HandleFunc("/", slackEventHandler)
	return r
}

// anonEventHandler handler for the /anonymize command
func anonEventHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.JeevesConfig{}
	cfg.GetConfig()

	verifier, err := slack.NewSecretsVerifier(
		r.Header,
		cfg.SigningToken,
	)
	if handleError(err) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
	err = commands.AnonymizeHandler(&cfg, r)
	if handleError(err) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// slackEventHandler The Event API handler for the slack APP
func slackEventHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.JeevesConfig{}
	cfg.GetConfig()

	check, err := jslack.VerifyWebHook(r, cfg.SigningToken)
	if handleError(err) || !check {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	body := buf.String()
	eventsAPIEvent, e := slackevents.ParseEvent(
		json.RawMessage(body),
		slackevents.OptionNoVerifyToken(),
	)
	if handleError(e) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if handleError(err) {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
		w.WriteHeader(http.StatusOK)
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// TODO deprecate this functionality
			_, _, _ = cfg.SlackApi.PostMessage(
				ev.Channel,
				slack.MsgOptionText("Anonymizing Repo", false),
			)
			url, candidateId, err := getLinkAndId(ev.Text)
			err = kubernetes.CreateAnonymizastionJob(&cfg, url, candidateId)
			if handleError(err) {
				_, _, _ = cfg.SlackApi.PostMessage(
					ev.Channel,
					slack.MsgOptionText("There was an Error Anonymizing Repo Please Contact Brendan", false),
				)
				w.WriteHeader(http.StatusOK)
				return
			}
			_, _, _ = cfg.SlackApi.PostMessage(
				ev.Channel,
				slack.MsgOptionText("API Anonymized", false),
			)

		}
	}
}

func handleError(err error) bool {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("webhook error occurred")
		return true
	}
	return false
}

func getLinkAndId(Message string) (string, string, error) {
	log.Printf("%v", Message)
	res := strings.Replace(Message, "<https://gitlab.com/", "", -1)
	res = strings.Replace(res, ">\u00a0", " ", -1)
	res = strings.Replace(res, "\u00a0", " ", -1)
	args := strings.Split(res, " ")
	log.Printf("%v", args)
	if len(args) != 3 {
		return "", "", fmt.Errorf("Wrong Format")
	}
	return args[1], args[2], nil
}
