package main

import (
	"fmt"
	"os/exec"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
)

func checkout(temp, repo, sha string) error {
	// don't clone the whole repo
	// it's too slow
	cmd := exec.Command("git", "clone", "--depth=100", "--recursive", "--branch=master", repo, temp)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	// checkout a commit (or branch or tag) of interest
	cmd = exec.Command("git", "checkout", "-qf", sha)
	cmd.Dir = temp
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}

	return nil
}

func build(temp, name string) error {
	cmd := exec.Command("docker", "build", "-t", name, ".")
	cmd.Dir = temp

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Running command failed: %s, %v", string(output), err)
	}
	return nil
}

func makeBinary(temp, image, name string, duration time.Duration) error {
	var (
		c   = make(chan error)
		cmd = exec.Command("docker", "run", "-t", "--privileged", "--name", name, "-v", path.Join(temp, "bundles")+":/go/src/github.com/docker/docker/bundles", image, "hack/make.sh", "binary cross")
	)
	cmd.Dir = temp

	go func() {
		output, err := cmd.CombinedOutput()
		if err != nil {
			// it's ok for the make command to return a non-zero exit
			// incase of a failed build
			if _, ok := err.(*exec.ExitError); !ok {
				log.Infof("Build failed: %s", string(output))
			} else {
				err = nil
			}
		}
		c <- err
	}()

	select {
	case err := <-c:
		if err != nil {
			return err
		}
	case <-time.After(duration):
		if err := cmd.Process.Kill(); err != nil {
			log.Infof("Killing process failed: %v", err)
		}
		return fmt.Errorf("Killed because build took to long")
	}
	return nil
}

func removeContainer(container string) {
	cmd := exec.Command("docker", "rm", "-f", container)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warnf("Removing container failed: %s, %v", string(output), err)
	}
}
