# Docker Binary Builder

Docker binary builder, triggered from nsq messages.

```console
$ docker-bb -h
Usage of docker-bb:
  -channel="binaries": nsq channel
  -d=false: run in debug mode
  -lookupd-addr="nsqlookupd:4161": nsq lookupd address
  -s3bucket="s3://test.docker.com/master/binaries/": s3 bucket to push binaries
  -s3region="us-east-1": s3 region where bucket lives
  -topic="hooks-docker": nsq topic
  -v=false: print version and exit (shorthand)
  -version=false: print version and exit
```

Example docker run command:

```bash
$ docker run -d --restart always \
    --link nsqlookupd1:nsqlookupd \
    -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /usr/local/bin/docker:/usr/local/bin/docker \
    -v /tmp:/tmp \
    -e DOCKER_HOST="unix:///var/run/docker.sock" \
    --privileged \
    --name binary-builder \
    dockercore/docker-bb -d -s3bucket="s3://jesss/test/docker/master/" \
    -s3region="us-west-1" \
    -topic hooks-docker -channel binaries \
    -lookupd-addr nsqlookupd:4161
```
