package create_test

import (
	"testing"

	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/repository/create"
	"github.com/stretchr/testify/require"
)

func TestCreateRepository(t *testing.T) {
	_, o := create.NewCmdCreateRepository()

	o.Kind = "fake"
	o.Server = "https://github.com"
	o.Token = "dummytoken"
	o.Username = "jstrachan"
	o.Owner = "myorg"
	o.Name = "myrepo"

	err := o.Run()
	require.NoError(t, err, "failed to create the repository")
	require.NotNil(t, o.Repository, "should have made a Repository")

	t.Logf("created repository %s", o.Repository.Link)
}
