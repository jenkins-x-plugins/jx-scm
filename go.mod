module github.com/jenkins-x-plugins/jx-scm

require (
	github.com/cpuguy83/go-md2man v1.0.10
	github.com/jenkins-x/go-scm v1.11.10
	github.com/jenkins-x/jx-helpers/v3 v3.0.90
	github.com/jenkins-x/jx-logging/v3 v3.0.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
)

replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

go 1.15
