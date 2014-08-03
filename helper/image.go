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

	"github.com/dustin/go-humanize"
)

type FetchConfig struct {
	URL          string
	Checksum     string
	ChecksumType string
	DownloadPath string
}

func FetchFile(config FetchConfig) (*os.File, error) {
	if config.URL == "" {
		panic("URL is required")
	}

	if config.Checksum == "" {
		panic("Checksum is required")
	}

	if config.ChecksumType == "" {
		panic("Checksum type is required")
	}

	if config.DownloadPath == "" {
		config.DownloadPath = os.TempDir()
	}

	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	_, filename := path.Split(u.Path)
	if filename == "" {
		filename = "unnamed"
	}

	filePath := filepath.Join(config.DownloadPath, filename)

	var file *os.File

	finfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsExist(err) && finfo.Size() > 0 {
			file, err = os.Open(filePath)
			if err != nil {
				return nil, err
			}
		} else if os.IsNotExist(err) || finfo.Size() == 0 {
			log.Printf("[DEBUG] %s file does not exist. Downloading it...", filename)

			data, err := download(config.URL)
			if err != nil {
				return nil, err
			}

			file, err = write(data, filePath)
			if err != nil {
				return nil, err
			}
			data.Close()
		} else {
			return nil, err
		}
	}

	// Makes sure the file is at the beginning so that verifying the checksum
	// does not fail
	file.Seek(0, 0)

	if err = VerifyChecksum(file, config.ChecksumType, config.Checksum); err != nil {
		log.Printf("[DEBUG] File on disk does not match current checksum.\n Downloading file again...")

		data, err := download(config.URL)
		if err != nil {
			return nil, err
		}

		file, err = write(data, filePath)
		if err != nil {
			return nil, err
		}
		data.Close()

		file.Seek(0, 0)

		if err = VerifyChecksum(file, config.ChecksumType, config.Checksum); err != nil {
			return nil, err
		}
	}

	return file, nil
}

func write(reader io.Reader, filePath string) (*os.File, error) {
	gzfile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	written, err := io.Copy(gzfile, reader)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] %s written to %s", humanize.Bytes(uint64(written)), filePath)

	return gzfile, nil
}

func download(URL string) (io.ReadCloser, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	resp, err := client.Get(URL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to fetch data, server returned code %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func UnpackFile(file io.Reader, destPath string) error {
	os.MkdirAll(destPath, 0740)

	//unzip
	log.Printf("[DEBUG] Unzipping file...")
	unzippedFile, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer unzippedFile.Close()

	//untar
	log.Printf("[DEBUG] Untaring archive...")
	return Untar(unzippedFile, destPath)
}
