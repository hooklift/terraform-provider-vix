package helper

import (
	"bytes"
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

	"github.com/dustin/go-humanize"
)

type Image struct {
	URL          string
	Checksum     string
	ChecksumType string
	DownloadPath string
}

func FetchImage(image Image) (*bytes.Buffer, error) {
	u, err := url.Parse(image.URL)
	if err != nil {
		return nil, err
	}

	_, filename := path.Split(u.Path)
	if filename == "" {
		filename = "unnamed"
	}

	filePath := filepath.Join(image.DownloadPath, filename)
	writeToDisk := false

	fstream := new(bytes.Buffer)

	finfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsExist(err) && finfo.Size() > 0 {
			file, err := os.Open(filePath)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			io.Copy(fstream, file)
		} else if os.IsNotExist(err) || finfo.Size() == 0 {
			log.Printf("[DEBUG] %s image does not exist. Downloading it...", filename)

			fstream, err = download(image.URL)
			if err != nil {
				return nil, err
			}

			writeToDisk = true
		} else {
			return nil, err
		}
	}

	buffer := bytes.NewReader(fstream.Bytes())
	if err = VerifyChecksum(buffer, image.ChecksumType, image.Checksum); err != nil {
		log.Printf("[DEBUG] Image in disk does not match current checksum.\n Downloading image again...")

		fstream.Reset()
		fstream, err = download(image.URL)
		if err != nil {
			return nil, err
		}

		buffer := bytes.NewReader(fstream.Bytes())
		if err = VerifyChecksum(buffer, image.ChecksumType, image.Checksum); err != nil {
			return nil, err
		}
		writeToDisk = true
	}

	if writeToDisk {
		gzfile, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}

		written, err := io.Copy(gzfile, fstream)
		if err != nil {
			return nil, err
		}
		defer gzfile.Close()

		log.Printf("[DEBUG] %s written to %s", humanize.Bytes(uint64(written)), filePath)
	}

	return fstream, nil
}

func download(imageURL string) (*bytes.Buffer, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to fetch data, server returned code %d", resp.StatusCode)
	}

	buffer := new(bytes.Buffer)
	io.Copy(buffer, resp.Body)

	return buffer, nil
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
