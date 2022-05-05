package create_pr_test

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"

	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/pull_request/create"
)

func TestCreatePullRequest(t *testing.T) {
	_, o := create_pr.NewCmdCreatePullRequest()

	o.Kind = "fake"
	o.Server = "https://github.com"
	o.Token = "dummytoken"
	o.Username = "WaciumaWanjohi"
	o.Owner = "myorg"
	o.Name = "myrepo"

	o.Head = "some_feature_branch"
	o.Base = "main"

	scmClient, err := o.Validate()
	assert.NoError(t, err)

	createPR(t, o, scmClient, err)
	createPRAgainWithoutAllowUpdate(t, o, scmClient, err)
	createPRAgainWithAllowUpdate(t, o, scmClient, err)
}

func createPR(t *testing.T, o *create_pr.Options, scmClient *scm.Client, err error) {
	fullName := scm.Join(o.Owner, o.Name)

	o.Title = "some pull request"
	o.Body = "Drastically improve the product"

	err = o.Run()
	assert.NoError(t, err, "failed to create the pull request")

	prs, _, err := scmClient.PullRequests.List(context.TODO(), fullName, scm.PullRequestListOptions{})
	assert.NoError(t, err, "failed to list pull requests")
	assert.Equal(t, 1, len(prs))
	assert.Equal(t, prs[0].Title, o.Title, "title not properly set")
	assert.Equal(t, prs[0].Body, o.Body, "body not properly set")
	assert.Equal(t, prs[0].Head.Ref, o.Head, "head not properly set")
	assert.Equal(t, prs[0].Base.Ref, o.Base, "base not properly set")
	assert.Equal(t, 1, prs[0].Number, "unexpected pr number set")
}

func createPRAgainWithoutAllowUpdate(t *testing.T, o *create_pr.Options, scmClient *scm.Client, err error) {
	fullName := scm.Join(o.Owner, o.Name)

	previousTitle := o.Title
	previousBody := o.Body

	o.Title = "Some new PR title"
	o.Body = "A reason for change we will not see"
	o.AllowUpdate = false

	err = o.Run()
	assert.Error(t, err, "expected pull request error did not occur")

	prs, _, err := scmClient.PullRequests.List(context.TODO(), fullName, scm.PullRequestListOptions{})
	assert.NoError(t, err, "failed to list pull requests")
	assert.Equal(t, 1, len(prs))
	assert.Equal(t, prs[0].Title, previousTitle, "title not properly set")
	assert.Equal(t, prs[0].Body, previousBody, "body not properly set")
	assert.Equal(t, prs[0].Head.Ref, o.Head, "head not properly set")
	assert.Equal(t, prs[0].Base.Ref, o.Base, "base not properly set")
	assert.Equal(t, 1, prs[0].Number, "unexpected pr number set")
}

func createPRAgainWithAllowUpdate(t *testing.T, o *create_pr.Options, scmClient *scm.Client, err error) {
	fullName := scm.Join(o.Owner, o.Name)

	o.Title = "An updated PR title"
	o.Body = "New reason for change"
	o.AllowUpdate = true

	err = o.Run()
	assert.NoError(t, err, "failed to update existing pull request")

	prs, _, err := scmClient.PullRequests.List(context.TODO(), fullName, scm.PullRequestListOptions{})
	assert.NoError(t, err, "failed to list pull requests")
	assert.Equal(t, 1, len(prs))
	assert.Equal(t, prs[0].Title, o.Title, "title not properly set")
	assert.Equal(t, prs[0].Body, o.Body, "body not properly set")
	assert.Equal(t, prs[0].Head.Ref, o.Head, "head not properly set")
	assert.Equal(t, prs[0].Base.Ref, o.Base, "base not properly set")
	assert.Equal(t, 1, prs[0].Number, "unexpected pr number set")
}