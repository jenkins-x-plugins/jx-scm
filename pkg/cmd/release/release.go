package release

import (
	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/release/update"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
)

// NewCmdRelease creates the new command
func NewCmdRelease() *cobra.Command {
	command := &cobra.Command{
		Use:     "release",
		Short:   "Commands for working with releases",
		Aliases: []string{"release"},
		Run: func(command *cobra.Command, args []string) {
			err := command.Help()
			if err != nil {
				log.Logger().Error(err.Error())
			}
		},
	}
	command.AddCommand(cobras.SplitCommand(update.NewCmdUpdateRelease()))
	return command
}
