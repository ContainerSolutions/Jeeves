package github

import (
	"context"
	"log"
	"net/http"

	"github.com/ContainerSolutions/jeeves/pkg/config"
	"github.com/google/go-github/v33/github"
	"github.com/slack-go/slack"
)

//IncomingWebhook handles an incoming webhook request
func IncomingWebhook(
	ctx context.Context,
	cfg config.JeevesConfig,
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
		err = PullRequestHandler(ctx, cfg, e, client)
	default:
		break
	}
	return err
}

func PullRequestHandler(
	ctx context.Context,
	cfg config.JeevesConfig,
	event *github.PullRequestEvent,
	client *github.Client,
) error {
	pr := event.PullRequest
	if event.GetAction() == "opened" && pr.Head.Repo.GetFork() {
		userName := pr.User.GetLogin()
		userUrl := pr.User.GetHTMLURL()
		pullRequestUrl := pr.GetHTMLURL()

		msg := (`A user with the Username: <` +
			userUrl + "|" +
			userName + `> Opened a <` +
			pullRequestUrl +
			`|Pull Request> on the API Excercise. I will let you know when they pass all the checks`)
		_, _, _ = cfg.SlackApi.PostMessage(
			cfg.SlackChannelID,
			slack.MsgOptionText(msg, false),
		)
	}
	return nil
}

func CheckPullRequests(client *github.Client) {
	ctx := context.Background()
	prs, _, err := client.PullRequests.List(
		ctx,
		"ContainerSolutions",
		"API-Excercise",
		&github.PullRequestListOptions{},
	)
	if err != nil {
		return
	}
	for _, pr := range prs {
		log.Printf("%v", pr.GetMergeableState())
		if pr.GetMergeable() {
			// Send a Message On The PR
			msg := "CONGRATULATIONS! Thank You for Submitting your API Excercise, If your email is on your Github Account and you fall are eligible according to our Terms and Conditions one of our Talent Team will reach out to you"
			err := createIssueComment(ctx, pr, client, msg)
			if err != nil {
				continue
			}
			// Kick Off A Job To Anonymize the Submission
			// Close PR
			err = closePullRequest(ctx, client, pr)
			if err != nil {
				continue
			}
		}
	}
}

func closePullRequest(
	ctx context.Context,
	client *github.Client,
	pr *github.PullRequest,
) error {
	updatedPr := &github.PullRequest{
		State: github.String("closed"),
	}
	_, _, err := client.PullRequests.Edit(
		ctx,
		pr.Base.User.GetLogin(),
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		updatedPr,
	)
	return err
}

// createIssueComment sends a comment to an issue/pull request
func createIssueComment(
	ctx context.Context,
	pr *github.PullRequest,
	client *github.Client,
	message string,
) error {
	comment := &github.IssueComment{Body: &message}
	_, _, err := client.Issues.CreateComment(
		ctx,
		pr.Base.Repo.Owner.GetLogin(),
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		comment,
	)
	return err
}
