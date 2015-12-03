FROM alpine
MAINTAINER Jessica Frazelle <jess@docker.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk update && apk add \
	ca-certificates \
	git \
	&& rm -rf /var/cache/apk/*

# make git happy
RUN	git config --global user.name docker-bb \
	&& git config --global user.email dockerbb@dockerproject.com \
	&& ln -s /.dockerinit /usr/bin/docker

COPY . /go/src/github.com/jfrazelle/docker-bb

RUN buildDeps=' \
		go \
		gcc \
		libc-dev \
		libgcc \
	' \
	set -x \
	&& apk update \
	&& apk add $buildDeps \
	&& cd /go/src/github.com/jfrazelle/docker-bb \
	&& go get -d -v github.com/jfrazelle/docker-bb \
	&& go build -o /usr/bin/docker-bb . \
	&& apk del $buildDeps \
	&& rm -rf /var/cache/apk/* \
	&& rm -rf /go \
	&& echo "Build complete."


ENTRYPOINT ["docker-bb"]
