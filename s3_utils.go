package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/crowdmob/goamz/s3"
)

func pushToS3(bucket *s3.Bucket, bucketpath, bundlesPath string) error {
	if _, err := os.Stat(bundlesPath); os.IsNotExist(err) {
		return fmt.Errorf("This is awkward, the bundles path DNE: %s", bundlesPath)
	}

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

		if err = uploadFileToS3(bucket, fpath, path.Join(bucketpath, relFilePath), ""); err != nil {
			return fmt.Errorf("Uploading %s to s3 failed: %v", fpath, err)
		}

		return nil
	}

	// walk the filepath
	if err := filepath.Walk(bundlesPath, walkFn); err != nil {
		return err
	}

	return nil
}

func uploadFileToS3(bucket *s3.Bucket, fpath, s3path, contentType string) error {
	contents, err := ioutil.ReadFile(fpath)
	if err != nil {
		return fmt.Errorf("Reading %q failed: %v", fpath, err)
	}

	// push the file to s3
	logrus.Debugf("Pushing %s to s3", s3path)
	if err := bucket.Put(s3path, contents, contentType, "public-read", s3.Options{CacheControl: "no-cache"}); err != nil {
		return err
	}
	logrus.Infof("Sucessfully pushed %s to s3", s3path)
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

// listFiles lists the files in a specific s3 bucket.
func listFiles(prefix, delimiter, marker string, maxKeys int, b *s3.Bucket) (files []s3.Key, err error) {
	resp, err := b.List(prefix, delimiter, marker, maxKeys)
	if err != nil {
		return nil, err
	}

	// append to files
	for _, fl := range resp.Contents {
		if strings.Contains(fl.Key, "index.html") || strings.Contains(fl.Key, "static") || strings.Contains(fl.Key, "logs") {
			continue
		}

		files = append(files, fl)
	}

	// recursion for the recursion god
	if resp.IsTruncated && resp.NextMarker != "" {
		f, err := listFiles(resp.Prefix, resp.Delimiter, resp.NextMarker, resp.MaxKeys, b)
		if err != nil {
			return nil, err
		}

		// append to files
		files = append(files, f...)
	}

	return files, nil
}
