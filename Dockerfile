FROM ghcr.io/jenkins-x/jx-boot:latest

RUN apk --no-cache add sed
    
COPY ./build/linux/jx-scm /usr/bin/jx-scm