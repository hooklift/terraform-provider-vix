package helper

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestUntar(t *testing.T) {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new tar archive.
	tw := tar.NewWriter(buf)

	// Add some files to the archive.
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling licence."},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Size: int64(len(file.Body)),
		}
		err := tw.WriteHeader(hdr)
		ok(t, err)

		_, err = tw.Write([]byte(file.Body))
		ok(t, err)
	}

	// Make sure to check the error on Close.
	err := tw.Close()
	ok(t, err)

	// Open the tar archive for reading.
	r := bytes.NewReader(buf.Bytes())
	destDir, err := ioutil.TempDir(os.TempDir(), "terraform-vix")
	ok(t, err)
	defer os.RemoveAll(destDir)

	_, err = Untar(r, destDir)
	ok(t, err)
}
