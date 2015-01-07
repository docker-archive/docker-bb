package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
)

const (
	VERSION = "v0.1.0"
)

var (
	lookupd string
	topic   string
	channel string
	bucket  string
	region  string
	debug   bool
	version bool
)

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&lookupd, "lookupd-addr", "nsqlookupd:4161", "nsq lookupd address")
	flag.StringVar(&topic, "topic", "hooks-docker", "nsq topic")
	flag.StringVar(&channel, "channel", "binaries", "nsq channel")
	flag.StringVar(&bucket, "s3bucket", "s3://master.dockerproject.com/", "s3 bucket to push binaries")
	flag.StringVar(&region, "s3region", "us-east-1", "s3 region where bucket lives")
	flag.Parse()
}

type Handler struct {
}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	hook, err := ParseHook(m.Body)
	if err != nil {
		// Errors will most likely occur because not all GH
		// hooks are the same format
		// we care about those that are pushing to master
		log.Debugf("Error parsing hook: %v", err)
		return nil
	}

	shortSha := hook.Sha[0:7]
	// checkout the code in a temp dir
	temp, err := ioutil.TempDir("", fmt.Sprintf("commit-%s", shortSha))
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	if err := checkout(temp, hook.Repo.CloneUrl, hook.Sha); err != nil {
		log.Warn(err)
		return err
	}
	log.Debugf("Checked out %s for %s", hook.Sha, hook.Repo.CloneUrl)

	var (
		image     = fmt.Sprintf("docker:commit-%s", shortSha)
		container = fmt.Sprintf("build-%s", shortSha)
	)
	log.Infof("image=%s container=%s\n", image, container)

	// build the image
	if err := build(temp, image); err != nil {
		log.Warn(err)
		return err
	}
	log.Debugf("Successfully built image %s", image)

	// make the binary
	defer removeContainer(container)
	if err = makeBinary(temp, image, container, 20*time.Minute); err != nil {
		log.Warn(err)
		return err
	}
	log.Debugf("Successfully built binaries for %s", hook.Sha)

	// read the version
	version, err := getBinaryVersion(temp)
	if err != nil {
		log.Warnf("Getting binary version failed: %v", err)
		return err
	}

	bundlesPath := path.Join(temp, "bundles", version, "cross")

	// create commit file
	if err := ioutil.WriteFile(path.Join(bundlesPath, "commit"), []byte(hook.Sha), 0755); err != nil {
		return err
	}

	// create version file
	if err := ioutil.WriteFile(path.Join(bundlesPath, "version"), []byte(version), 0755); err != nil {
		return err
	}

	// push to s3
	if err = pushToS3(bundlesPath); err != nil {
		log.Warn(err)
		return err
	}
	return nil
}

func main() {
	// set log level
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	bb := &Handler{}
	if err := ProcessQueue(bb, QueueOptsFromContext(topic, channel, lookupd)); err != nil {
		log.Fatal(err)
	}
}
