package create

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x-plugins/jx-scm/pkg/scmclient"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
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

		# creates a new git repository using a URL
		%s repository create --git-kind gitlab https://mygitserver/myowner/myrepo
	`)

	info = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	options.BaseOptions
	scmclient.Options

	Args        []string
	Owner       string
	Name        string
	Description string
	HomePage    string
	Template    string
	GitPushHost string
	Private     bool
	Confirm     bool
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
			o.Args = args
			err := o.Run()
			helper.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&o.Owner, "owner", "o", "", "the owner of the repository to create. Either an organisation or username.  For Azure, include the project: 'organization/project'")
	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "the name of the repository to create")
	cmd.Flags().StringVarP(&o.Description, "description", "d", "", "the repository description")
	cmd.Flags().StringVarP(&o.HomePage, "home-page", "", "", "the repository home page")
	cmd.Flags().StringVarP(&o.Template, "template", "", "", "the git template repository to create the repository from")
	cmd.Flags().StringVarP(&o.GitPushHost, "push-host", "", "", "the git host to use when pushing to the git repository. Only really useful in BDD tests if using something like 'kubectl portforward' to access a git repository where you want to push from outside the cluster with a different host name to the host name used inside the cluster")
	cmd.Flags().BoolVarP(&o.Private, "private", "", false, "if the repository should be private")
	cmd.Flags().BoolVarP(&o.Confirm, "confirm", "", false, "confirms creating the repository")

	o.Options.AddFlags(cmd)
	o.BaseOptions.AddBaseFlags(cmd)
	return cmd, o
}

// Validate create parameters
func (o *Options) Validate() (*scm.Client, error) {
	if len(o.Args) > 0 {
		repoURL := o.Args[0]
		u, err := url.Parse(repoURL)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse repository URL %s", repoURL)
		}
		if o.Server != "" {
			return nil, errors.Errorf("specified --server when already supplied %s", repoURL)
		}
		if o.Owner != "" {
			return nil, errors.Errorf("specified --owner when already supplied %s", repoURL)
		}
		if o.Name != "" {
			return nil, errors.Errorf("specified --name when already supplied %s", repoURL)
		}

		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, "/")
		path = strings.TrimSuffix(path, ".git")
		names := strings.Split(path, "/")

		if len(names) < 2 {
			return nil, errors.Errorf("repository URL should be in the form https://myserver/myowner/myrepo but was %s", repoURL)
		}

		o.Name = names[len(names)-1]
		o.Owner = names[len(names)-2]

		remainingNames := names[0 : len(names)-2]
		u.Path = "/" + stringhelpers.UrlJoin(remainingNames...)
		o.Server = strings.TrimSuffix(u.String(), "/")
	}

	err := o.BaseOptions.Validate()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate base options")
	}
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

	if o.Template != "" {
		err = o.createTemplate(o.Template)
		if err != nil {
			return errors.Wrapf(err, "failed to create template")
		}
	}
	return nil
}

func (o *Options) createTemplate(template string) error {
	g := o.GitClient
	dir, err := gitclient.CloneToDir(g, template, "")
	if err != nil {
		return errors.Wrapf(err, "failed to clone template %s", template)
	}
	remote := "newrepo"

	cloneURL := o.Repository.Clone
	if cloneURL != "" && o.GitPushHost != "" {
		u, err := url.Parse(cloneURL)
		if err != nil {
			return errors.Wrapf(err, "failed to parse repository clone URL %s", cloneURL)
		}
		u.Host = o.GitPushHost
		cloneURL = u.String()
		log.Logger().Infof("switching to the git clone URL %s", info(cloneURL))
	}

	username := o.Options.Username
	remoteURL, err := stringhelpers.URLSetUserPassword(cloneURL, username, o.Options.Token)
	if err != nil {
		return errors.Wrapf(err, "failed to create the remote git URL for %s and user %s", cloneURL, username)
	}
	_, err = g.Command(dir, "remote", "add", remote, remoteURL)
	if err != nil {
		return errors.Wrapf(err, "failed to add remote %s %s", remote, cloneURL)
	}

	branch, err := gitclient.Branch(g, dir)
	if err != nil {
		return errors.Wrapf(err, "failed to get the current branch in dir %s", dir)
	}

	_, err = g.Command(dir, "push", remote, branch)
	if err != nil {
		return errors.Wrapf(err, "failed to push remote %s branch %s to %s", remote, branch, cloneURL)
	}

	log.Logger().Infof("pushed the template repository %s to %s", info(template), info(cloneURL))
	return nil
}
