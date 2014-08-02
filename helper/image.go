package helper

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

type Image struct {
	URL          string
	Checksum     string
	ChecksumType string
	DownloadPath string
}

func FetchImage(image Image) (*os.File, error) {
	u, err := url.Parse(image.URL)
	if err != nil {
		return nil, err
	}

	_, filename := path.Split(u.Path)
	if filename == "" {
		filename = "default.tar.gz"
	}

	filePath := filepath.Join(image.DownloadPath, filename)

	// client := NewHttpClient()
	writeToDisk := false

	var file io.Reader

	finfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsExist(err) && finfo.Size() > 0 {
			file, err = os.Open(filePath)
			if err != nil {
				return nil, err
			}
		} else if os.IsNotExist(err) || finfo.Size() == 0 {
			log.Printf("[DEBUG] %s image does not exist. Downloading it...", filename)
			// data, err := client.GetRetry(image.URL)
			file, err = download(image.URL)
			if err != nil {
				return nil, err
			}

			//file = bytes.NewBuffer(data)
			writeToDisk = true
		} else {
			return nil, err
		}
	}

	if err = VerifyChecksum(file, image.ChecksumType, image.Checksum); err != nil {
		log.Printf("[DEBUG] Image in disk does not match current checksum.\n Downloading image again...")

		// data, err := client.GetRetry(image.URL)
		// if err != nil {
		// 	return nil, err
		// }
		file, err := download(image.URL)
		if err != nil {
			return nil, err
		}

		//file = bytes.NewBuffer(data)

		if err = VerifyChecksum(file, image.ChecksumType, image.Checksum); err != nil {
			return nil, err
		}
		writeToDisk = true
	}

	if writeToDisk {
		gzfile, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}

		written, err := io.Copy(gzfile, file)
		if err != nil {
			return nil, err
		}
		log.Printf("[DEBUG] %d bytes written to %s", written, filePath)

		file = gzfile
	}

	return file.(*os.File), nil
}

func download(imageURL string) (io.Reader, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to fetch data, server returned code %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func UnpackImage(file io.Reader, destPath string) error {
	os.MkdirAll(destPath, 0740)

	//unzip
	log.Printf("[DEBUG] Unzipping image...")
	unzippedFile, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer unzippedFile.Close()

	//untar
	log.Printf("[DEBUG] Untaring image...")
	return Untar(unzippedFile, destPath)
}
