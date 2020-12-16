package router

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ContainerSolutions/jeeves/pkg/commands"
	"github.com/ContainerSolutions/jeeves/pkg/config"
	jslack "github.com/ContainerSolutions/jeeves/pkg/slack"
	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// GetRouter returns the routes for the server
func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/github/events", githubEventHandler)
	r.HandleFunc("/slack/events", slackEventHandler)
	r.HandleFunc("/anonymize", anonEventHandler)
	return r
}

// anonEventHandler handler for the /anonymize command
func anonEventHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.JeevesConfig{}
	cfgErr := cfg.GetConfig()
	if handleError(cfgErr) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	verifier, err := slack.NewSecretsVerifier(
		r.Header,
		cfg.SlackSigningToken,
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

// githubEventHandler The Event API handler for the github APP
func githubEventHandler(w http.ResponseWriter, r *http.Request) {}

// slackEventHandler The Event API handler for the slack APP
func slackEventHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.JeevesConfig{}
	cfgErr := cfg.GetConfig()
	if handleError(cfgErr) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	check, err := jslack.VerifyWebHook(r, cfg.SlackSigningToken)
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
	// If the Slack Events API sends a Challenge it will be handled by this
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
		switch innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			w.WriteHeader(http.StatusOK)
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
