# Jeeves The Talent Bot

This is the Slack Bot that the Container Solutions Talent Team Uses to Anonymize Candidates Git Repos.

## Deployment

Currently it is deployed into the CS Engineering Shared Cluster Manually

## How It Works

The bot just listens to mentions on a slack channel and then creates a Kubernetes Job within the cluster that handles the anonyimization, the anonymizer is just a bash script that can be found in the [anonymizer directory](/anonymizer)

In the channel you would just run

```
@Jeeves https://gitlab.com/<gitlab user name>/api-exercise <candidate id>
```

## Roadmap

- Add Tests
- Make Commands more robust
- Automate Build and deployment
- Add creation of Scoring Google Sheet
- To Be Determined
