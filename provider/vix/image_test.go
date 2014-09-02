package vix

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadTARGZ(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-tar")
		w.Header().Set("Content-Encoding", "x-gzip")

		file, err := os.Open("./fixtures/test.box")
		ok(t, err)
		assert(t, file != nil, "Failed loading fixture file")
		defer file.Close()

		io.Copy(w, file)
	}))
	defer ts.Close()

	image := Image{
		URL:          ts.URL,
		Checksum:     "35fd19dc1bb7e18a365c1c589df2292942c197a4",
		ChecksumType: "sha1",
	}

	destDir, err := ioutil.TempDir(os.TempDir(), "terraform-vix")
	ok(t, err)
	defer os.RemoveAll(destDir)

	err = image.Download(destDir)
	ok(t, err)

	filename := image.file.Name()
	assert(t, filename != "", fmt.Sprintf("%v == %v", filename, nil))
	finfo, err := image.file.Stat()
	ok(t, err)

	size := finfo.Size()
	assert(t, size > 0, fmt.Sprintf("Image file is empty: %d", size))
}

func TestUnpack(t *testing.T) {
	gzipFile, err := os.Open("./fixtures/test.box")
	ok(t, err)

	image := Image{
		file: gzipFile,
	}

	destDir, err := ioutil.TempDir(os.TempDir(), "terraform-vix")
	ok(t, err)
	defer os.RemoveAll(destDir)

	_, err = image.Unpack(destDir)
	ok(t, err)

	files, err := ioutil.ReadDir(destDir)
	ok(t, err)
	assert(t, len(files) == 2, "There should be two files inside tgz package")
}

func TestVerifyChecksum(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "terraform-vix-test")
	ok(t, err)

	file.WriteString("Test")

	var tests = []struct {
		file      *os.File
		algorithm string
		checksum  string
	}{
		// echo -n Test | gsha256sum
		{file, "sha256", "532eaabd9574880dbf76b9b8cc00832c20a6ec113d682299550d7a6e0f345e25"},
		// echo -n Test | gsha1sum
		{file, "sha1", "640ab2bae07bedc4c163f679a746f7ab7fb5d1fa"},
		// echo -n Test | gsha512sum
		{file, "sha512", "c6ee9e33cf5c6715a1d148fd73f7318884b41adcb916021e2bc0e800a5c5dd97f5142178f6ae88c8fdd98e1afb0ce4c8d2c54b5f37b30b7da1997bb33b0b8a31"},
		// echo -n Test | gmd5sum
		{file, "md5", "0cbc6611f5540bd0809a388dc95a615b"},
	}

	for _, test := range tests {
		image := Image{
			Checksum:     test.checksum,
			ChecksumType: test.algorithm,
			file:         test.file,
		}
		ok(t, image.verify())
	}
}
