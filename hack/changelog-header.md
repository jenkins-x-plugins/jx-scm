### Linux

```shell
curl -L https://github.com/jenkins-x-plugins/jx-scm/releases/download/v{{.Version}}/jx-scm-linux-amd64.tar.gz | tar xzv 
sudo mv jx-scm /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x-plugins/jx-scm/releases/download/v{{.Version}}/jx-scm-darwin-amd64.tar.gz | tar xzv
sudo mv jx-scm /usr/local/bin
```

