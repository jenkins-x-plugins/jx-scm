package cmd

import (
	pull "github.com/jenkins-x-plugins/jx-scm/pkg/cmd/pr"
	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/release"
	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/repository"
	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/version"
	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/spf13/cobra"
)

// Main creates the new command
func Main() *cobra.Command {
	// Initializing logger
	log.Logger()
	cmd := &cobra.Command{
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: rootcmd.TopLevelCommand,
		},
		Short: "GitOps utility commands",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Logger().Error(err.Error())
			}
		},
	}
	cmd.AddCommand(pull.NewCmdPullRequest())
	cmd.AddCommand(release.NewCmdRelease())
	cmd.AddCommand(repository.NewCmdRepository())

	cmd.AddCommand(cobras.SplitCommand(version.NewCmdVersion()))
	return cmd
}
