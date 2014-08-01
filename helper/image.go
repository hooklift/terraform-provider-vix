package helper

import (
	"compress/gzip"
	"io/ioutil"
	"os"
)

func FetchImage(url, checksum, checksumType, dest string) error {
	client := NewHttpClient()
	data, err := client.Get(url)
	if err != nil {
		return err
	}

	err = VerifyChecksum(data, checksumType, checksum)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, data, 0740)
	if err != nil {
		return err
	}

	return nil
}

func UnpackImage(path, dest string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = os.MkdirAll(dest, 0740)
	if err != nil {
		return err
	}

	//unzip
	data, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	//untar
	err = Untar(data, dest)
	if err != nil {
		return err
	}

	return nil
}
