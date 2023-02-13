package update

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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cmdLong = templates.LongDesc(`
		Update a release
`)

	cmdExample = templates.Examples(`
		# updates a release to change the title
		%s release update --owner foo --repository bar --tag v1.2.3 --title something

		# updates a release to make it not a pre-release
		%s release update --owner foo --repository bar --tag v1.2.3 --pre-release false
	`)

	_ = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	scmclient.Options

	Owner       string
	Name        string
	Title       string
	Description string
	Tag         string
	PreRelease  bool
	ScmClient   *scm.Client
}

// NewCmdUpdateRelease updates a release
func NewCmdUpdateRelease() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Updates a release",
		Long:    cmdLong,
		Example: fmt.Sprintf(cmdExample, rootcmd.BinaryName, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.Options.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.Owner, "owner", "o", "", "the owner of the repository to update. Either an organisation or username.  For Azure, include the project: 'organization/project'")
	cmd.Flags().StringVarP(&o.Name, "name", "r", "", "the name of the repository to update")
	cmd.Flags().StringVarP(&o.Tag, "tag", "", "", "the tag of the release to update")

	cmd.Flags().StringVarP(&o.Description, "description", "", "", "the updated release description")
	cmd.Flags().StringVarP(&o.Title, "title", "", "", "the updated release title")
	cmd.Flags().BoolVarP(&o.PreRelease, "prerelease", "", true, "the updated prerelease status, true to identify the release as a prerelease, false to identify the release as a full release.")
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
	if o.Tag == "" {
		return nil, options.MissingOption("tag")
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

	releaseInput := &scm.ReleaseInput{
		Description: o.Description,
		Title:       o.Title,
		Prerelease:  o.PreRelease,
		Tag:         o.Tag,
	}

	release, _, err := scmClient.Releases.FindByTag(ctx, fullName, o.Tag)
	if err != nil {
		return errors.Wrapf(err, "failed to find release %s %s", fullName, o.Tag)
	}
	_, _, err = scmClient.Releases.Update(ctx, fullName, release.ID, releaseInput)
	if err != nil {
		return errors.Wrapf(err, "failed to update release %s %s, id: %v", fullName, o.Tag, release.ID)
	}
	return nil
}
