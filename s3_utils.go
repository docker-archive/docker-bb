package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"github.com/rakyll/magicmime"
)

func pushToS3(bundlesPath string) error {
	if _, err := os.Stat(bundlesPath); os.IsNotExist(err) {
		return fmt.Errorf("This is awkward, the bundles path DNE: %s", bundlesPath)
	}

	// use env variables to connect to s3
	auth, err := aws.EnvAuth()
	if err != nil {
		return fmt.Errorf("AWS Auth failed: %v", err)
	}

	// connect to s3 bucket
	s := s3.New(auth, aws.GetRegion(region))
	bucketname, bucketpath := bucketParts(bucket)
	bucket := s.Bucket(bucketname)

	//walk the bundles directory
	walkFn := func(fpath string, info os.FileInfo, err error) error {
		stat, err := os.Stat(fpath)
		if err != nil {
			return err
		}

		relFilePath, err := filepath.Rel(bundlesPath, fpath)
		if err != nil || (fpath == bundlesPath && stat.IsDir()) {
			// Error getting relative path OR we are looking
			// at the root path. Skip in both situations.
			return nil
		}

		if stat.IsDir() {
			return nil
		}

		if err = uploadFileToS3(bucket, fpath, path.Join(bucketpath, relFilePath)); err != nil {
			log.Warnf("Uploading %s to s3 failed: %v", fpath, err)
			return err
		}
		return nil
	}

	err = filepath.Walk(bundlesPath, walkFn)
	return err
}

func uploadFileToS3(bucket *s3.Bucket, fpath, s3path string) error {
	// try to get the mime type
	mimetype := ""
	mm, err := magicmime.New(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	if err != nil {
		log.Debugf("Magic meme failed for: %v", err)
	} else {
		mimetype, err = mm.TypeByFile(fpath)
		if err != nil {
			log.Debugf("Mime type detection for %s failed: %v", fpath, err)
		}
	}

	contents, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Warnf("Reading %q failed: %v", fpath, err)
	}

	// push the file to s3
	log.Debugf("Pushing %s to s3", s3path)
	if err := bucket.Put(s3path, contents, mimetype, "", s3.Options{}); err != nil {
		return err
	}
	log.Infof("Sucessfully pushed %s to s3", s3path)
	return nil
}

// parse for the parts of the bucket name
func bucketParts(bucket string) (bucketname, path string) {
	s3Prefix := "s3://"
	if strings.HasPrefix(bucket, s3Prefix) {
		bucket = strings.Replace(bucket, s3Prefix, "", 1)
	}
	parts := strings.SplitN(bucket, "/", 2)

	if len(parts) <= 1 {
		path = ""
	} else {
		path = parts[1]
	}
	return parts[0], path
}
