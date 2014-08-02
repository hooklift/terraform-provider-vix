package helper

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func FetchImage(URL, checksum, checksumType string) (*os.File, error) {
	client := NewHttpClient()
	client.Timeout = time.Duration(15) * time.Second
	data, err := client.GetRetry(URL)
	if err != nil {
		return nil, err
	}

	log.Print("[DEBUG] Verifying checksum for image...")
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
	log.Printf("[DEBUG] Unzipping image...")
	data, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	//untar
	log.Printf("[DEBUG] Untaring image...")
	return Untar(data, destDir)
}
