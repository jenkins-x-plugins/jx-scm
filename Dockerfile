FROM gcr.io/jenkinsxio/jx-cli-base:latest

ENTRYPOINT ["jx-scm"]

COPY ./build/linux/jx-scm /usr/bin/jx-scm