package close_pr

import (
	"context"
	"fmt"

	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x-plugins/jx-scm/pkg/scmclient"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
	`)

	info = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	scmclient.Options

	Owner string
	Name  string

	PR        int
	Before    int
	Size      int
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
	return cmd, o
}

// Run transforms the YAML files
func (o *Options) Validate() (*scm.Client, error) {
	scmClient, err := o.Options.Validate()
	if err != nil {
		return scmClient, errors.Wrapf(err, "failed to validate options")
	}

	if o.Owner == "" {
		return nil, options.MissingOption("owner")
	}
	if o.Name == "" {
		return nil, options.MissingOption("repository")
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

	// check at least one flags (pr and before) are set but not both
	if o.PR < 0 && o.Before < 0 {
		return errors.New("please set --pr or --before flag")
	}

	if o.PR > 0 && o.Before > 0 {
		return errors.New("please set ony one of --pr or --before flags")
	}
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

	return nil
}
