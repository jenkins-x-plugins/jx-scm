package update_test

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"

	"github.com/jenkins-x-plugins/jx-scm/pkg/cmd/release/update"
)

func TestUpdateRelease(t *testing.T) {
	_, o := update.NewCmdUpdateRelease()

	o.Kind = "fake"
	o.Server = "https://github.com"
	o.Token = "dummytoken"
	o.Username = "rawlingsj"
	o.Owner = "myorg"
	o.Name = "myrepo"
	o.Tag = "v9.9.9"

	scmClient, err := o.Validate()
	assert.NoError(t, err)

	input := &scm.ReleaseInput{
		Title:       "foo",
		Description: "bar",
		Tag:         o.Tag,
		Prerelease:  true,
	}
	fullName := scm.Join(o.Owner, o.Name)

	_, _, err = scmClient.Releases.Create(context.TODO(), fullName, input)
	assert.NoError(t, err, "failed to create the release")

	o.Title = "wine"
	o.Description = "cheese"
	o.PreRelease = false
	err = o.Run()
	assert.NoError(t, err, "failed to update the release")

	release, _, err := scmClient.Releases.FindByTag(context.TODO(), fullName, o.Tag)
	assert.NoError(t, err, "failed to find the updated release")
	assert.Equal(t, "wine", release.Title, "title should have been updated")
	assert.Equal(t, "cheese", release.Description, "description should have been updated")
	assert.Equal(t, false, release.Prerelease, "prerelease should have been updated")
}
