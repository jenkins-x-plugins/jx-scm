package repository

import (
	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/repository/clone"
	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/repository/create"
	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/repository/remove"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
)

// NewCmdRepository creates the new command
func NewCmdRepository() *cobra.Command {
	command := &cobra.Command{
		Use:     "repository",
		Short:   "Commands for working with source repositories",
		Aliases: []string{"repo", "repos", "repositories"},
		Run: func(command *cobra.Command, args []string) {
			err := command.Help()
			if err != nil {
				log.Logger().Errorf(err.Error())
			}
		},
	}
	command.AddCommand(cobras.SplitCommand(clone.NewCmdCloneRepository()))
	command.AddCommand(cobras.SplitCommand(create.NewCmdCreateRepository()))
	command.AddCommand(cobras.SplitCommand(remove.NewCmdRemoveRepository()))
	return command
}
