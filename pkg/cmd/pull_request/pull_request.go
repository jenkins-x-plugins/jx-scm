package pull_request

import (
	close_pr "github.com/jenkins-x-plugins/jx-scm/pkg/cmd/pull_request/close"
	create_pr "github.com/jenkins-x-plugins/jx-scm/pkg/cmd/pull_request/create"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
)

// NewCmdPullRequest creates the new command
func NewCmdPullRequest() *cobra.Command {
	command := &cobra.Command{
		Use:     "pull-request",
		Short:   "Commands for working with pull-requests",
		Aliases: []string{"pr"},
		Run: func(command *cobra.Command, args []string) {
			err := command.Help()
			if err != nil {
				log.Logger().Errorf(err.Error())
			}
		},
	}
	command.AddCommand(cobras.SplitCommand(close_pr.NewCmdClosePullRequest()))
	command.AddCommand(cobras.SplitCommand(create_pr.NewCmdCreatePullRequest()))

	return command
}
