package scmclient

import (
	"context"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Options common CLI arguments for working with a git server
type Options struct {
	Kind      string
	Server    string
	Username  string
	Token     string
	ScmClient *scm.Client
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Kind, "kind", "k", "", "the kind of git server to use")
	cmd.Flags().StringVarP(&o.Server, "server", "s", "", "the git server URL to use")
	cmd.Flags().StringVarP(&o.Username, "username", "u", "", "the user name to use on the git server")
	cmd.Flags().StringVarP(&o.Token, "token", "t", "", "the token to use on the git server")

}

// Validate validates the options and returns the ScmClient
func (o *Options) Validate() (*scm.Client, error) {
	if o.Kind == "" {
		return nil, options.MissingOption("kind")
	}
	if o.Server == "" {
		return nil, options.MissingOption("server")
	}
	if o.Token == "" {
		return nil, options.MissingOption("token")
	}

	var err error
	if o.ScmClient == nil {
		o.ScmClient, err = factory.NewClient(o.Kind, o.Server, o.Token)
		if err != nil {
			return o.ScmClient, errors.Wrapf(err, "failed to create ScmClient for kind %s server %s", o.Kind, o.Server)
		}
	}

	ctx := context.Background()
	if o.Username == "" {
		user, _, err := o.ScmClient.Users.Find(ctx)
		if err != nil {
			return o.ScmClient, errors.Wrapf(err, "failed to find current user")
		}
		o.Username = user.Login
	}
	return o.ScmClient, nil
}
