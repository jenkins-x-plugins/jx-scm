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

	fullName := scm.Join(o.Owner, o.Name)

	scmClient, err := o.Validate()
	assert.NoError(t, err)

	createPR(t, o, fullName, scmClient, err)
	createPRAgainWithoutAllowUpdate(t, o, fullName, scmClient, err)
	createPRAgainWithAllowUpdate(t, o, fullName, scmClient, err)
}

func createPR(t *testing.T, o *create_pr.Options, fullName string, scmClient *scm.Client, err error) {
	o.Title = "some pull request"
	o.Body = "Drastically improve the product"

	err = o.Run()
	assert.NoError(t, err, "failed to create the pull request")

	prs := assertExpectedPRFound(t, o, fullName, scmClient, err)

	assert.Equal(t, prs[0].Title, o.Title, "title not properly set")
	assert.Equal(t, prs[0].Body, o.Body, "body not properly set")
}

func createPRAgainWithoutAllowUpdate(t *testing.T, o *create_pr.Options, fullName string, scmClient *scm.Client, err error) {
	previousTitle := o.Title
	previousBody := o.Body

	o.Title = "Some new PR title"
	o.Body = "A reason for change we will not see"
	o.AllowUpdate = false

	err = o.Run()
	assert.Error(t, err, "expected pull request error did not occur")

	prs := assertExpectedPRFound(t, o, fullName, scmClient, err)

	assert.Equal(t, prs[0].Title, previousTitle, "title not properly set")
	assert.Equal(t, prs[0].Body, previousBody, "body not properly set")
}

func createPRAgainWithAllowUpdate(t *testing.T, o *create_pr.Options, fullName string, scmClient *scm.Client, err error) {
	o.Title = "An updated PR title"
	o.Body = "New reason for change"
	o.AllowUpdate = true

	err = o.Run()
	assert.NoError(t, err, "failed to update existing pull request")

	prs := assertExpectedPRFound(t, o, fullName, scmClient, err)

	assert.Equal(t, prs[0].Title, o.Title, "title not properly set")
	assert.Equal(t, prs[0].Body, o.Body, "body not properly set")
}

func assertExpectedPRFound(t *testing.T, o *create_pr.Options, fullName string, scmClient *scm.Client, err error) []*scm.PullRequest {
	prs, _, err := scmClient.PullRequests.List(context.TODO(), fullName, &scm.PullRequestListOptions{})
	assert.NoError(t, err, "failed to list pull requests")
	assert.Equal(t, 1, len(prs))
	assert.Equal(t, prs[0].Head.Ref, o.Head, "head not properly set")
	assert.Equal(t, prs[0].Base.Ref, o.Base, "base not properly set")
	assert.Equal(t, 1, prs[0].Number, "unexpected pr number set")
	return prs
}
