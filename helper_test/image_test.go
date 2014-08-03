package helper_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/c4milo/terraform_vix/helper"
)

func TestFetch(t *testing.T) {
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

	destDir, err := ioutil.TempDir(os.TempDir(), "terraform-vix")
	ok(t, err)
	//defer os.RemoveAll(destDir)

	fetchConfig := helper.FetchConfig{
		URL:          ts.URL,
		Checksum:     "35fd19dc1bb7e18a365c1c589df2292942c197a4",
		ChecksumType: "sha1",
		DownloadPath: destDir,
	}

	file, err := helper.FetchFile(fetchConfig)
	ok(t, err)

	assert(t, file != nil, fmt.Sprintf("%v == %v", file, nil))
	finfo, err := file.Stat()
	ok(t, err)

	assert(t, finfo.Size() > 0, fmt.Sprintf("Image file is empty: %d", finfo.Size()))
}

func TestFetchingExistingFile(t *testing.T) {

}

func TestUnpackFile(t *testing.T) {
	gzipFile, err := os.Open("./fixtures/test.box")
	ok(t, err)

	destDir, err := ioutil.TempDir(os.TempDir(), "terraform-vix")
	ok(t, err)
	defer os.RemoveAll(destDir)

	err = helper.UnpackFile(gzipFile, destDir)
	ok(t, err)
}
