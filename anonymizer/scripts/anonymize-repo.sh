#!/bin/bash
set -e

export FILTER_BRANCH_SQUELCH_WARNING=1
user=$1
user_id=$2
dir=$(echo $1 | sed 's/\//-/g')
keyfile=$CS_REVIEWER_KEY
# Clone Repo
echo "######################################"
echo "Cloning Repo"
echo "######################################"
git clone git@gitlab.com:$user.git $dir

# This will loop through all branches on the remote and sync them with local

echo "######################################"
echo "Tracking All Branches"
echo "######################################"
command cd $dir
for branch in `git branch -a | grep remotes | grep -v HEAD | grep -v master`;
do
    git branch --track ${branch#remotes/origin/} $branch
done
command cd ..

# Copy the repo to a new destination and cd
command cp -r $dir $user_id
command cd $user_id

# Rewrite history with anonymized user data
# We may be doing multiple rewrites, so we must force subsequent ones.
# We're throwing away the backups anyway.
echo "######################################"
echo "Anonymizing Repo"
echo "######################################"
    command git filter-branch -f \
            --msg-filter \
            'sed "s/Signed-off-by:.*>//g" |  sed "s/See merge request.*//g"' \
            --env-filter \
            'export GIT_AUTHOR_NAME="Anonymous Candidate"
            export GIT_AUTHOR_EMAIL="anon@repo.com"
            export GIT_COMMITTER_NAME="Anonymous Candidate"
            export GIT_COMMITTER_EMAIL="anon@repo.com"' \
            --tag-name-filter cat -- --branches --tags
# Remove remote references
    for r in `git remote`; do git remote rm $r; done

echo "######################################"
echo "Cleaning Up"
echo "######################################"
command cd ..
command zip -r $user_id.zip $user_id
command rm -rf $user_id/
command rm -rf $dir
echo "######################################"
echo "Uploading Repo"
echo "######################################"
gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS
gsutil cp $user_id.zip gs://anonymized-repos/$user_id.zip
