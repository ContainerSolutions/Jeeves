# Jeeves The Talent Bot

This is the Slack Bot that the Container Solutions Talent Team Uses to Anonymize Candidates Git Repos.

## Secrets

Secrets are encrypted with Google KMS, you will need access to [the keys](https://console.cloud.google.com/security/kms/key/manage/global/jeeves/jeeves?project=cs-engineering-256009) in order to decrypt them

### Decrypting Secrets

```bash
./scripts/decrypt
```

### Encrypting Secrets

```bash
./scripts/encrypt
```

## Deployment

Currently it is deployed into the CS Engineering Shared Cluster Manually

## How It Works

The bot just listens to mentions on a slack channel and then creates a Kubernetes Job within the cluster that handles the anonyimization, the anonymizer is just a bash script that can be found in the [anonymizer directory](/anonymizer)

In the channel you would just run

```
/anonymize https://gitlab.com/<gitlab user name>/api-exercise <candidate id>
```

## Roadmap

- Automate Build and deployment
- Add creation of Scoring Google Sheet
- Add Github Intergrations
