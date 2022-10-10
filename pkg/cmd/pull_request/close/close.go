package close_pr

import (
	"context"
	"fmt"

	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x-plugins/jx-scm/pkg/scmclient"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	create_pr "github.com/jenkins-x-plugins/jx-scm/pkg/cmd/pull_request/create"
)

var (
	cmdLong = templates.LongDesc(`
		Update a release
`)

	cmdExample = templates.Examples(`
		# closes pull requests foo/bar number 123
		%s pull-request close --owner foo --name bar --pr 123

		# closes all open pull requests on foo/bar before pull request number 200
		%s pull-request close --owner foo --name bar --before 200

		# close an open pull request on foo/bar from branch baz onto base branch main
		%s pull-request close --owner foo --name bar --head baz --base main
	`)

	info = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	scmclient.Options

	Owner string
	Name  string

	PR     int
	Before int
	Size   int

	Head string
	Base string

	ScmClient *scm.Client
}

// NewCmdClosePullRequest closes a pull request
func NewCmdClosePullRequest() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "close",
		Short:   "closes a pull request",
		Long:    cmdLong,
		Example: fmt.Sprintf(cmdExample, rootcmd.BinaryName, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.Options.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.Owner, "owner", "o", "", "the owner of the repository that contains pull requests to close. Either an organisation or username")
	cmd.Flags().StringVarP(&o.Name, "name", "r", "", "the name of the repository that contains pull requests to close")

	cmd.Flags().IntVarP(&o.PR, "pr", "", 0, "the pull request to close")
	cmd.Flags().IntVarP(&o.Size, "size", "", 200, "the number of open pull requests to return if using --before, defaults to 200")
	cmd.Flags().IntVarP(&o.Before, "before", "", 0, "a pull request number to used to close ALL open pull requests before it")
	cmd.Flags().StringVarP(&o.Head, "head", "", "", "the name of the branch where changes are implemented")
	cmd.Flags().StringVarP(&o.Base, "base", "", "main", "the name of the branch the changes would be pulled into")

	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("name")

	return cmd, o
}

// Validate validates the options and returns the ScmClient
func (o *Options) Validate() (*scm.Client, error) {
	// check at least one flags (pr and before) are set but not both
	var prFlagSet, beforeFlagSet, baseOrHeadFlagSet int
	if o.PR < 0 {
		prFlagSet = 1
	}

	if o.Before < 0 {
		beforeFlagSet = 1
	}

	if o.Head != "" || o.Base != "" {
		baseOrHeadFlagSet = 1
	}

	if prFlagSet+beforeFlagSet+baseOrHeadFlagSet != 1 {
		return nil, errors.New("must set either --pr or --before or both --head and -- base flags")
	}

	if (o.Head == "" && o.Base != "") || (o.Head != "" && o.Base == "") {
		return nil, errors.New("--base and --head must be set together")
	}

	scmClient, err := o.Options.Validate()
	if err != nil {
		return scmClient, errors.Wrapf(err, "failed to validate options")
	}

	return scmClient, nil
}

// Run transforms the YAML files
func (o *Options) Run() error {
	scmClient, err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate options")
	}

	fullName := scm.Join(o.Owner, o.Name)

	ctx := context.Background()

	// if pr flag set then close it
	if o.PR > 0 {
		log.Logger().Infof("closing pull request%s %v", fullName, o.PR)
		_, err := scmClient.PullRequests.Close(ctx, fullName, o.PR)
		if err != nil {
			return errors.Wrapf(err, "failed to close pull request %s #%v", fullName, o.PR)
		}
	}

	if o.Before > 0 {
		// if before then first list open pull requests
		pullRequests, _, err := scmClient.PullRequests.List(ctx, fullName, &scm.PullRequestListOptions{Open: true, Size: o.Size})
		if err != nil {
			return errors.Wrapf(err, "failed to list pull requests for #%s", fullName)
		}
		// loop over all  open PRs and close any that are before --before value
		for _, pr := range pullRequests {
			if pr.Number < o.Before {
				_, err := scmClient.PullRequests.Close(ctx, fullName, pr.Number)
				if err != nil {
					return errors.Wrapf(err, "failed to close pull request %s #%v", fullName, pr.Number)
				}
				log.Logger().Infof("closing pull request%s %v", fullName, pr.Number)
			}
		}
	}

	if o.Head != "" && o.Base != "" {
		foundOpenPR, pullRequestNumber := create_pr.FindOpenPullRequestByBranches(o.Head, o.Base, scmClient, ctx, fullName)
		if !foundOpenPR {
			log.Logger().Infof("no open pull request from branch %s to base branch %s", o.Head, o.Base)
		} else {
			log.Logger().Infof("closing pull request #%d", pullRequestNumber)
			_, err := scmClient.PullRequests.Close(ctx, fullName, pullRequestNumber)
			if err != nil {
				return errors.Wrapf(err, "failed to close pull request %s #%v", fullName, o.PR)
			}
		}
	}

	return nil
}
