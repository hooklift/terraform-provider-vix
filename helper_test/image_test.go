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

func TestFetchImage(t *testing.T) {
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

	file, err := helper.FetchImage(ts.URL, "de147793e322837247c7b83f2f96033918825d4b68fe16dc4e7242fac7015611", "sha256")
	ok(t, err)
	assert(t, file != nil, fmt.Sprintf("%v == %v", file.Name(), nil))
	file.Close()
}

func TestUnpackImage(t *testing.T) {
	gzipFile, err := os.Open("./fixtures/test.box")
	ok(t, err)

	destDir, err := ioutil.TempDir(os.TempDir(), "terraform-vix")
	ok(t, err)
	defer os.RemoveAll(destDir)

	err = helper.UnpackImage(gzipFile, destDir)
	ok(t, err)
}
