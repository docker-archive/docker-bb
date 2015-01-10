FROM golang:latest
MAINTAINER Jess Frazelle <jess@docker.com>

RUN apt-get update && apt-get install -y \
    libmagic-dev \
    --no-install-recommends

RUN go get github.com/bitly/go-nsq && \
    go get github.com/Sirupsen/logrus && \
    go get github.com/crowdmob/goamz/aws && \
    go get github.com/rakyll/magicmime && \
    go get github.com/drone/go-github/github

COPY . /go/src/github.com/jfrazelle/docker-bb
RUN cd /go/src/github.com/jfrazelle/docker-bb && go install . ./...
ENV PATH $PATH:/go/bin

# make git happy
RUN git config --global user.name docker-bb && \
    git config --global user.email dockerbb@dockerproject.com && \
    ln -s /.dockerinit /usr/bin/docker

ENTRYPOINT ["docker-bb"]
