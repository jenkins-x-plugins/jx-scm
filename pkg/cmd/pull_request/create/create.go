package create_pr

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x-plugins/jx-scm/pkg/scmclient"
)

var (
	cmdLong = templates.LongDesc(`
		Creates a pull request in the given repository, requesting the head branch be merged into the base branch
`)

	cmdExample = templates.Examples(`
		# creates a pull request for a branch 
		%s pull-request create \
			--owner foo \
			--name bar \
			--title "chore: a good reason to merge" \
			--body "Useful details for reviewers" \
			--head some-feature-branch \
			--base main

		# if pull request from head branch to base exists, updates the pr. otherwise creates the pr
		%s pull-request create \
			--owner foo \
			--name bar \
			--title "chore: a good reason to merge" \
			--body "Some new reasons to merge" \
			--head some-feature-branch \
			--base main \
			--allow-update
	`)

	info = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	scmclient.Options

	Owner string
	Name  string

	Title string
	Body  string
	Head  string
	Base  string

	AllowUpdate bool

	ScmClient *scm.Client
}

// NewCmdCreatePullRequest creates a pull request
func NewCmdCreatePullRequest() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Creates a pull request",
		Long:    cmdLong,
		Example: fmt.Sprintf(cmdExample, rootcmd.BinaryName, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.Options.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.Owner, "owner", "o", "", "the owner of the repository. Either an organisation or username")
	cmd.Flags().StringVarP(&o.Name, "name", "r", "", "the name of the repository")


	cmd.Flags().StringVarP(&o.Title, "title", "", "", "the title of the new pull request")
	cmd.Flags().StringVarP(&o.Body, "body", "", "", "the contents of the pull request")
	cmd.Flags().StringVarP(&o.Head, "head", "", "", "the name of the branch where your changes are implemented")
	cmd.Flags().StringVarP(&o.Base, "base", "", "main", "the name of the branch you want the changes pulled into")

	cmd.Flags().BoolVarP(&o.AllowUpdate, "allow-update", "", false, "if an open pull request from head branch to base branch exists, setting flag to true will update the pull request")

	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("head")

	return cmd, o
}

// Validate validates the options and returns the ScmClient
func (o *Options) Validate() (*scm.Client, error) {
	scmClient, err := o.Options.Validate()
	if err != nil {
		return scmClient, errors.Wrapf(err, "failed to validate options")
	}

	return scmClient, nil
}

// Run implements the command
func (o *Options) Run() error {
	scmClient, err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate options")
	}

	fullName := scm.Join(o.Owner, o.Name)

	ctx := context.Background()

	pullRequestInput := &scm.PullRequestInput{
		Title: o.Title,
		Body:  o.Body,
		Head:  o.Head,
		Base:  o.Base,
	}

	shouldUpdate, existingPullRequestNumber := updateNecessary(o.Head, o.Base, o.AllowUpdate, scmClient, ctx, fullName)

	if shouldUpdate {
		res, _, err := scmClient.PullRequests.Update(ctx, fullName, existingPullRequestNumber, pullRequestInput)
		if err != nil {
			return errors.Wrapf(err, "failed to update existing pull request #%d in repo '%s'", existingPullRequestNumber, fullName)
		}

		log.Logger().Infof("updated pull request #%d in repo '%s'. url: %s", res.Number, res.Base.Repo.FullName, res.Link)

		return nil
	}

	res, _, err := scmClient.PullRequests.Create(ctx, fullName, pullRequestInput)
	if err != nil {
		return errors.Wrapf(err, "failed to create a pull request in the repository '%s' with the title '%s'", fullName, o.Title)
	}

	log.Logger().Infof("created pull request #%d in repo '%s'. url: %s", res.Number, res.Base.Repo.FullName, res.Link)

	return nil
}

func updateNecessary(head string, base string, updateAllowed bool, scmClient *scm.Client, ctx context.Context, fullName string) (bool, int) {
	if !updateAllowed {
		return false, 0
	}
	pullRequestListOptions := scm.PullRequestListOptions{Page: 1, Size: 5000, Open: true, Closed: false}

	openPullRequests, _, err := scmClient.PullRequests.List(ctx, fullName, &pullRequestListOptions)
	if err != nil {
		log.Logger().Errorf("listing pull requests in repo '%s' failed: %s", fullName, err)
		return false, 0
	}

	for _, openPullRequest := range openPullRequests {
		if openPullRequest.Head.Ref == head && openPullRequest.Base.Ref == base {
			return true, openPullRequest.Number
		}
	}

	return false, 0
}
