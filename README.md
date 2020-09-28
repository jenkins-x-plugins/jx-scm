# jx-scm

[![Documentation](https://godoc.org/github.com/jenkins-x-plugins/jx-scm?status.svg)](https://pkg.go.dev/mod/github.com/jenkins-x-plugins/jx-scm)
[![Go Report Card](https://goreportcard.com/badge/github.com/jenkins-x-plugins/jx-scm)](https://goreportcard.com/report/github.com/jenkins-x-plugins/jx-scm)
[![Releases](https://img.shields.io/github/release-pre/jenkins-x/jx-scm.svg)](https://github.com/jenkins-x-plugins/jx-scm/releases)
[![LICENSE](https://img.shields.io/github/license/jenkins-x/jx-scm.svg)](https://github.com/jenkins-x-plugins/jx-scm/blob/master/LICENSE)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://slack.k8s.io/)

jx-scm is a small command line tool working with git providers using [go-scm](https://github.com/jenkins-x/go-scm)

## Getting Started

Download the [jx-scm binary](https://github.com/jenkins-x-plugins/jx-scm/releases) for your operating system and add it to your `$PATH`.

There will be an `app` you can install soon too...

## Commands

See the [jx-scm command reference](docs/cmd/jx-scm.md#see-also)


## Developing

If you wish to work on a local clone of [go-scm](https://github.com/jenkins-x/go-scm) then:

```bash                  
git clone https://github.com/jenkins-x/go-scm
```                                          

Then in the local `go.mod` file add the following at the end:


``` 
replace github.com/jenkins-x/go-scm  => PathToTheAboveGitClone
```                                                           

You can now build this repository using your local modifications and try the locally built binary in `build/jx-scm` or run the unit tests via `make test`