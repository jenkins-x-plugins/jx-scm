package remove

import (
	"context"
	"fmt"
	"time"

	"github.com/jenkins-x-plugins/jx-scm/pkg/rootcmd"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input"
	"github.com/jenkins-x/jx-helpers/v3/pkg/input/survey"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/stringhelpers"
	"github.com/jenkins-x/jx-helpers/v3/pkg/termcolor"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cmdLong = templates.LongDesc(`
		Removes one or more repositories
`)

	cmdExample = templates.Examples(`
		# removes all the repositories in the owner with the given filter
		%s repository remove --owner myuser -f mything

		# removes all the repositories in the owner created before the given time
		%s repository remove --owner myuser --created-before '02 Jan 06 15:04 MST'

		# removes all the repositories in the owner created 30 days ago
		%s repository remove --owner myuser --created-days-ago 30  --confirm
	`)

	info = termcolor.ColorInfo
)

// LabelOptions the options for the command
type Options struct {
	scmhelpers.Factory

	Owner             string
	Name              string
	Includes          []string
	Excludes          []string
	CreatedBefore     string
	CreatedDaysAgo    int
	Confirm           bool
	DryRun            bool
	FailOnRemoveError bool
	Input             input.Interface
	CreatedBeforeTime *time.Time
}

// NewCmdCreateRepository creates a command object for the command
func NewCmdRemoveRepository() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "Removes one or more repositories",
		Aliases: []string{"delete", "rm"},
		Long:    cmdLong,
		Example: fmt.Sprintf(cmdExample, rootcmd.BinaryName, rootcmd.BinaryName, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.Factory.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.Owner, "owner", "o", "", "the owner of the repository to create. Either an organisation or username")
	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "the name of the repository to create")
	cmd.Flags().StringVarP(&o.CreatedBefore, "created-before", "", "", "the time expression for removing repositories created before this time")
	cmd.Flags().IntVarP(&o.CreatedDaysAgo, "created-days-ago", "", 0, "remove repositories created more than this number of days ago")
	cmd.Flags().StringArrayVarP(&o.Includes, "filter", "f", nil, "the text filter to match the name")
	cmd.Flags().StringArrayVarP(&o.Excludes, "exclude", "x", nil, "the text filter to exclude")
	cmd.Flags().BoolVarP(&o.Confirm, "confirm", "", false, "confirms the removal without prompting the user")
	cmd.Flags().BoolVarP(&o.DryRun, "dry-run", "", false, "disables actually deleting the repository so you can test the filtering")
	cmd.Flags().BoolVarP(&o.FailOnRemoveError, "fail-on-error", "", false, "stops removing repositories if a remove failsg")
	return cmd, o
}

// Run transforms the YAML files
func (o *Options) Validate() (*scm.Client, error) {
	if o.Factory.GitServerURL == "" {
		o.Factory.GitServerURL = giturl.GitHubURL
	}
	scmClient, err := o.Factory.Create()
	if err != nil {
		return scmClient, errors.Wrapf(err, "failed to create SCM client")
	}

	if o.Owner == "" {
		return nil, options.MissingOption("owner")
	}

	if o.Input == nil {
		o.Input = survey.NewInput()
	}

	if o.CreatedDaysAgo > 0 {
		if o.CreatedBefore != "" {
			return nil, errors.Errorf("you cannot supply --created-before and --created-days-ago")
		}

		hours := time.Duration(-24 * o.CreatedDaysAgo)
		t := time.Now().Add(hours * time.Hour)
		o.CreatedBeforeTime = &t
	}

	if o.CreatedBefore != "" {
		t, err := time.Parse(time.RFC822, o.CreatedBefore)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse created-before time %s", o.CreatedBefore)
		}
		o.CreatedBeforeTime = &t
	}
	return scmClient, nil
}

// Run transforms the YAML files
func (o *Options) Run() error {
	scmClient, err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate options")
	}

	ctx := context.Background()

	user, _, err := scmClient.Users.Find(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to lookup current user")
	}

	currentUser := user.Login

	listOptions := scm.ListOptions{
		Size: 100,
	}
	for {
		var repos []*scm.Repository
		var resp *scm.Response
		if o.Owner == currentUser {
			repos, resp, err = scmClient.Repositories.List(ctx, listOptions)
			if err != nil && !scmhelpers.IsScmNotFound(err) {
				if resp == nil {
					return errors.Wrapf(err, "failed to list user repositories - no response")
				}
				return errors.Wrapf(err, "failed to list user repositories status %d", resp.Status)
			}
		} else {
			repos, resp, err = scmClient.Repositories.ListOrganisation(ctx, o.Owner, listOptions)
			if err != nil && !scmhelpers.IsScmNotFound(err) {
				if resp == nil {
					return errors.Wrapf(err, "failed to list user repositories - no response")
				}
				return errors.Wrapf(err, "failed to list organisation repositories %s status %d", o.Owner, resp.Status)
			}
		}
		if len(repos) == 0 {
			break
		}

		var deleteRepos []string
		for _, repo := range repos {
			if o.Matches(repo) {
				deleteRepos = append(deleteRepos, repo.FullName)
			}
		}

		for _, name := range deleteRepos {
			if o.DryRun {
				log.Logger().Infof("would remove repository %s", info(name))
				continue
			}

			if !o.Confirm {
				flag, err := o.Input.Confirm("do you want to delete repository "+name+"?", false, "confirm you wish to remove the repository")
				if err != nil {
					return errors.Wrapf(err, "failed to confirm removal")
				}

				if !flag {
					log.Logger().Infof("not removing repository %s", info(name))
					continue
				}
			}
			resp, err = scmClient.Repositories.Delete(ctx, name)
			if err != nil {
				if resp == nil {
					if o.FailOnRemoveError {
						return errors.Wrapf(err, "failed to delete repository %s no status", name)
					}
					log.Logger().Warnf("failed to delete repository %s no status", name)
				} else {
					if o.FailOnRemoveError {
						return errors.Wrapf(err, "failed to delete repository %s status %d", name, resp.Status)
					}
					log.Logger().Warnf("failed to delete repository %s status %d", name, resp.Status)
				}
			}
			log.Logger().Infof("removed repository %s", info(name))
		}

		listOptions.Page++
	}
	return nil
}

// Matches returns true if the repository matches the filter
func (o *Options) Matches(repo *scm.Repository) bool {
	if repo.Namespace != o.Owner {
		return false
	}
	if o.CreatedBeforeTime != nil && o.CreatedBeforeTime.Before(repo.Created) {
		return false
	}
	if len(o.Includes) == 0 {
		return true
	}
	return stringhelpers.StringContainsAny(repo.Name, o.Includes, o.Excludes)
}
