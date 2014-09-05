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

func TestDownload(t *testing.T) {
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
