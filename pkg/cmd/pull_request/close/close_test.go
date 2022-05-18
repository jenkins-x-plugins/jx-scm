package close_pr_test

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"

	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/pull_request/close"
)

func TestClosePullRequestByNumber(t *testing.T) {
	// Needs testing
}

func TestClosePullRequestByBefore(t *testing.T) {
	// Needs testing
}

func TestClosePullRequestByBranches(t *testing.T) {
	_, o := close_pr.NewCmdClosePullRequest()

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

	initialExistingPR := &scm.PullRequestInput{
		Title: "some-title",
		Head:  "some_feature_branch",
		Base:  "main",
		Body:  "some information about this PR",
	}

	_, _, err = scmClient.PullRequests.Create(context.TODO(), fullName, initialExistingPR)
	assert.NoError(t, err, "failed to pre-create pull requests")

	err = o.Run()
	assert.NoError(t, err, "failed to close the pull request")

	prs, _, err := scmClient.PullRequests.List(context.TODO(), fullName, &scm.PullRequestListOptions{Open: true})
	assert.NoError(t, err, "failed to list pull requests")
	assert.Equal(t, 0, len(prs))

	err = o.Run()
	assert.NoError(t, err, "failed to close the pull request")
}
