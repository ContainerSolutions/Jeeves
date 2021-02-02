package github

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/ContainerSolutions/jeeves/pkg/config"
	"github.com/ContainerSolutions/jeeves/pkg/kubernetes"
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

func CheckPullRequests(
	cfg config.JeevesConfig,
	client *github.Client,
) {
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
		// List does not populate all pull request fields hence we have to individually fetch the pull requests
		checks, _, _ := client.Checks.ListCheckRunsForRef(
			ctx,
			"ContainerSolutions",
			"API-Excercise",
			pr.Head.GetSHA(),
			&github.ListCheckRunsOptions{},
		)
		var checkStatus string
		for _, check := range checks.CheckRuns {
			//TODO: Figure out how to make this less dependant on the checks name
			if check.GetName() == "run_api_test" {
				checkStatus = check.GetConclusion()
				log.Printf("%+v", checkStatus)
			}
		}

		fork := false
		fail := false

		if pr.Base.Repo.GetFullName() != pr.Head.Repo.GetFullName() {
			fork = true
		}

		// Check Restricted Files Not Changed
		if fork {
			files, _, _ := client.PullRequests.ListFiles(
				ctx,
				"ContainerSolutions",
				"API-Excercise",
				pr.GetNumber(),
				&github.ListOptions{},
			)
			for _, file := range files {
				if strings.Contains(file.GetFilename(), ".github") || strings.Contains(file.GetFilename(), ".ci") {
					msg := "PLEASE NOTE: You are not able to make changes in the `.github` or `.ci` directories"
					_ = createIssueComment(ctx, pr, client, msg)
					// Close PR
					err = closePullRequest(ctx, client, pr)
					if err != nil {
						continue
					}
					fail = true
					break
				}
			}
		}

		if fork && checkStatus == "success" && !fail {
			// Send a Message On The PR
			msg := "CONGRATULATIONS! Thank You for Submitting your API Excercise, If your email is on your Github Account and you are eligible according to our Terms and Conditions one of our Talent Team will reach out to you"
			err := createIssueComment(ctx, pr, client, msg)
			if err != nil {
				continue
			}
			// Kick Off A Job To Anonymize the Submission
			candidateId := pr.Head.User.GetLogin() + "-" + pr.GetUpdatedAt().String()
			candidateId = strings.ReplaceAll(candidateId, " ", "-")
			candidateId = strings.ReplaceAll(candidateId, "+", "-")
			candidateId = strings.ReplaceAll(candidateId, ":", "-")
			err = kubernetes.CreateAnonymizastionJob(&cfg, "github.com", pr.Head.Repo.GetFullName(), candidateId)
			if err != nil {
				continue
			}
			// Close PR
			err = closePullRequest(ctx, client, pr)
			if err != nil {
				continue
			}
			// Send Message to Slack
			userName := pr.User.GetLogin()
			userUrl := pr.User.GetHTMLURL()
			msg = (`A user with the Username: <` +
				userUrl + "|" +
				userName + `> the submission can be found in the Google Storage Bucket with The following ID: ` + candidateId)
			_, _, _ = cfg.SlackApi.PostMessage(cfg.SlackChannelID, slack.MsgOptionText(msg, false))
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
