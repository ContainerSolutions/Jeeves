package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v32/github"
)

//IncomingWebhook handles an incoming webhook request
func IncomingWebhook(
	ctx context.Context,
	r *http.Request,
	payload []byte,
	client *github.Client,
) error {
	event, parseErr := github.ParseWebHook(github.WebHookType(r), payload)
	if parseErr != nil {
		return parseErr
	}
	var err error
	switch e := event.(type) {
	case *github.PullRequestEvent:
		err = PullRequestHandler(ctx, e, client)
	default:
		break
	}
	return err
}

func PullRequestHandler(
	ctx context.Context,
	event *github.PullRequestEvent,
	client *github.Client,
) error {
	log.Printf("%v", event)
	return nil
}
