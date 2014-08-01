package helper

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
)

func FetchImage(URL, checksum, checksumType string) (*os.File, error) {
	client := NewHttpClient()
	data, err := client.Get(URL)
	if err != nil {
		return nil, err
	}

	err = VerifyChecksum(data, checksumType, checksum)
	if err != nil {
		return nil, err
	}

	file, err := ioutil.TempFile(os.TempDir(), "terraform-vix")
	if err != nil {
		return nil, err
	}

	io.Copy(file, bytes.NewReader(data))

	return file, nil
}

func UnpackImage(file *os.File, destDir string) error {
	os.MkdirAll(destDir, 0740)

	//unzip
	data, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	//untar
	return Untar(data, destDir)
}
