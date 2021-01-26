package clone

import (
	"fmt"
	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/cli"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cmdLong = templates.LongDesc(`
		Clones a git repository
`)

	cmdExample = templates.Examples(`
		# creates a new git repository in the given server
		%s repository clone https://myserver.com/myowner/myrepo
	`)

	info = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	options.BaseOptions

	Args             []string
	CloneURL         string
	GitClient        gitclient.Interface
	GitCommandRunner cmdrunner.CommandRunner
}

// NewCmdCloneRepository creates a command object for the command
func NewCmdCloneRepository() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "clone",
		Short:   "Clones a git repository",
		Long:    cmdLong,
		Example: fmt.Sprintf(cmdExample, rootcmd.BinaryName, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			o.Args = args
			err := o.Run()
			helper.CheckErr(err)
		},
	}

	o.BaseOptions.AddBaseFlags(cmd)
	return cmd, o
}

// Run transforms the YAML files
func (o *Options) Validate() error {
	err := o.BaseOptions.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate base options")
	}

	if len(o.Args) <= 0 {
		return errors.Errorf("missing repository URL argument")
	}
	o.CloneURL = o.Args[0]

	if o.GitCommandRunner == nil {
		o.GitCommandRunner = cmdrunner.QuietCommandRunner
	}
	if o.GitClient == nil {
		o.GitClient = cli.NewCLIClient("", o.GitCommandRunner)
	}
	return nil
}

// Run transforms the YAML files
func (o *Options) Run() error {
	err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate options")
	}
	g := o.GitClient
	dir := "."
	_, err = g.Command(dir, "clone", o.CloneURL)
	if err != nil {
		return errors.Wrapf(err, "failed to clone repository %s", o.CloneURL)
	}
	log.Logger().Infof("cloned the repository %s", info(o.CloneURL))
	return nil
}
