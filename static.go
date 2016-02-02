package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/crowdmob/goamz/s3"
	units "github.com/docker/go-units"
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

		<p>These binaries are built and updated with each commit to the master branch of Docker. Want to use that cool new feature that was just merged? Download your system's binary and check out the master docs at <a href="http://docs.master.dockerproject.com" target="_blank">docs.master.dockerproject.com</a>.</p>

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
			{{ range $key, $value := . }}
				<tr>
					<td valign="top"><a href="/{{ $value.Key }}"><img src="/static/{{ $value.Key | ext }}.png" alt="[ICO]"/></a></td>
					<td><a href="/{{ $value.Key }}">{{ $value.Key }}</a></td>
					<td>{{ $value.Size | size }}</td>
					<td>{{ $value.LastModified }}</td>
				</tr>
			{{ end }}
            </tbody>
        </table>
    </div>
</body>
</html>`
)

// create the index.html file
func createIndexFile(bucket *s3.Bucket, bucketpath string) error {
	// list all the files
	files, err := listFiles(bucketpath, bucketpath, "", 2000, bucket)
	if err != nil {
		return fmt.Errorf("Listing all files in bucket failed: %v", err)
	}

	// create a temp file for the index
	tmp, err := ioutil.TempFile("", "index.html")
	if err != nil {
		return fmt.Errorf("Creating temp file failed: %v", err)
	}
	defer os.RemoveAll(tmp.Name())

	// set up custom functions
	funcMap := template.FuncMap{
		"ext": func(name string) string {
			if strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".md5") {
				return "text"
			}
			return "default"
		},
		"size": func(s int64) string {
			return units.HumanSize(float64(s))
		},
	}

	// parse & execute the template
	tmpl, err := template.New("").Funcs(funcMap).Parse(index)
	if err != nil {
		return fmt.Errorf("Parsing template failed: %v", err)
	}

	if err := tmpl.Execute(tmp, files); err != nil {
		return fmt.Errorf("Execute template failed: %v", err)
	}

	// push the file to s3
	if err = uploadFileToS3(bucket, tmp.Name(), path.Join(bucketpath, "index.html")); err != nil {
		return fmt.Errorf("Uploading %s to s3 failed: %v", tmp.Name(), err)
	}

	return nil
}
