# Api Review Anonymizer

## In order to use you will need access to the following

- A service account that has access to push to the [api reviewer bucket](https://console.cloud.google.com/storage/browser/anonymized-repos;tab=objects?forceOnBucketsSortingFiltering=false&organizationId=879351307558&project=api-reviews-289606&prefix=&forceOnObjectsSortingFiltering=false)
- The SSH Key for the cs-reviewer user in order to clone the API Excercise

## How to use

To run the below it is assumed that both the service account and the ssh key are in the keys directory

```bash
export GITLAB_REPO=<gitlab user>/API-Exercise
export ANONYMIZED_ID=1234

docker run --rm \
    -v  $PWD/keys:/infra/.user/.ssh
    -e GOOGLE_APPLICATION_CREDENTIALS=/infra/.user/.ssh/credentials.json
    gcr.io/api-reviews-289606/anonymizer:latest $GITLAB_REPO $ANONYMIZED_ID
```
> Note the `ssh key` has to be volumed into the directory `/infra/.user/.ssh` as `id_rsa`, the ` GOOGLE_APPLICATION_CREDENTIALS` can be anywhere as long as you set the environment variable accordingly
