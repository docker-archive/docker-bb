package main

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/crowdmob/goamz/s3"
)

const (
	index string = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <title>Docker Master Binaries</title>
    <link rel="stylesheet" href="/static/style.css" />
</head>
<body>
    <h1>Docker Master Binaries</h1>
    <div class="wrapper">
        <table>
            <thead>
                <tr>
                    <th><img src="/static/folder.png" alt="[ICO]"/></th>
                    <th>Name</th>
                    <th>Size</th>
                    <th>Uploaded Date</th>
                </tr>
            </thead>
            <tbody>
			{{ . }}
            </tbody>
        </table>
    </div>
</body>
</html>`
)

func createIndexFile(bucket *s3.Bucket, html string) error {
	path := "index.html"
	contents := strings.Replace(index, "{{ . }}", html, -1)

	// push the file to s3
	log.Debugf("Pushing %s to s3", path)
	if err := bucket.Put(path, []byte(contents), "text/html", "public-read", s3.Options{}); err != nil {
		return err
	}
	log.Infof("Sucessfully pushed %s to s3", path)

	return nil
}
