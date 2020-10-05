package create

import (
	"context"
	"fmt"

	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x-plugins/jx-scm/pkg/scmclient"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cmdLong = templates.LongDesc(`
		Creates a new git provider in a git server
`)

	cmdExample = templates.Examples(`
		# creates a new git repository in the given server
		%s repository create --git-kind gitlab --git-server https://myserver.com --owner myuser --name myrepo
	`)

	info = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	scmclient.Options

	Owner       string
	Name        string
	Description string
	HomePage    string
	Private     bool
	Repository  *scm.Repository
}

// NewCmdCreateRepository creates a command object for the command
func NewCmdCreateRepository() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Creates a new git provider in a git server",
		Long:    cmdLong,
		Example: fmt.Sprintf(cmdExample, rootcmd.BinaryName, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.Options.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.Owner, "owner", "o", "", "the owner of the repository to create. Either an organisation or username")
	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "the name of the repository to create")
	cmd.Flags().StringVarP(&o.Description, "description", "d", "", "the repository description")
	cmd.Flags().StringVarP(&o.HomePage, "home-page", "", "", "the repository home page")
	cmd.Flags().BoolVarP(&o.Private, "private", "", false, "if the repository should be private")
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
		return nil, options.MissingOption("name")
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

	repoInput := &scm.RepositoryInput{
		Name:        o.Name,
		Description: o.Description,
		Homepage:    o.HomePage,
		Private:     o.Private,
	}
	if o.Username != o.Owner {
		repoInput.Namespace = o.Owner
	}
	o.Repository, _, err = scmClient.Repositories.Create(ctx, repoInput)
	if err != nil {
		return errors.Wrapf(err, "failed to create repository %s", fullName)
	}

	log.Logger().Infof("created repository %s at %s", info(fullName), info(o.Repository.Link))
	return nil
}
