package vix

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

// A virtual machine image definition
type Image struct {
	// Image URL where to download from
	URL string
	// Checksum of the image, used to check integrity after downloading it
	Checksum string
	// Algorithm use to check the checksum
	ChecksumType string
	// Password to decrypt the virtual machine if it is encrypted. This is used by
	// VIX to be able to open the virtual machine
	Password string
	// Internal file reference
	file *os.File
}

// Downloads and a virtual machine image
func (img *Image) Download(destPath string) error {
	if img.URL == "" {
		panic("URL is required")
	}

	if img.Checksum == "" {
		panic("Checksum is required")
	}

	if img.ChecksumType == "" {
		panic("Checksum type is required")
	}

	if destPath == "" {
		destPath = os.TempDir()
	}

	u, err := url.Parse(img.URL)
	if err != nil {
		return err
	}

	_, filename := path.Split(u.Path)
	if filename == "" {
		filename = "unnamed"
	}

	os.MkdirAll(destPath, 0740)

	filePath := filepath.Join(destPath, filename)

	fetchAndWrite := func() error {
		data, err := img.fetch(img.URL)
		if err != nil {
			return err
		}

		img.file, err = img.write(data, filePath)
		if err != nil {
			return err
		}
		data.Close()

		return nil
	}

	log.Printf("[DEBUG] Opening %s...", filePath)
	img.file, err = os.Open(filePath)
	if err != nil {
		log.Printf("[DEBUG] %s file does not exist. Downloading it...", filename)

		if err = fetchAndWrite(); err != nil {
			return err
		}
	}

	if err = img.verify(); err != nil {
		log.Printf("[DEBUG] File on disk does not match current checksum.\n Downloading file again...")

		if err = fetchAndWrite(); err != nil {
			return err
		}

		if err = img.verify(); err != nil {
			return err
		}
	}

	return nil
}

// Gets a VM image through HTTP
func (img *Image) fetch(URL string) (io.ReadCloser, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	fragments, err := url.Parse(URL)

	if err == nil && fragments.Scheme == "file" {
		filePath := fragments.Path
		_, err := os.Stat(filePath)
		if err != nil {
			return nil, err
		}
		t := &http.Transport{}
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		client = &http.Client{Transport: t}
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

// Writes the downloading stream down to a file
func (img *Image) write(reader io.Reader, filePath string) (*os.File, error) {
	log.Printf("[DEBUG] Downloading file data to %s", filePath)

	compressedFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	written, err := io.Copy(compressedFile, reader)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] %s written to %s", humanize.Bytes(uint64(written)), filePath)

	return compressedFile, nil
}

// Verifies the image package integrity after it is downloaded
func (img *Image) verify() error {
	// Makes sure the file cursor is positioned at the beginning of the file
	_, err := img.file.Seek(0, 0)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Verifying image checksum...")
	var hasher hash.Hash

	switch img.ChecksumType {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		return fmt.Errorf("[ERROR] Crypto algorithm no supported: %s", img.ChecksumType)
	}
	_, err = io.Copy(hasher, img.file)
	if err != nil {
		return err
	}

	result := fmt.Sprintf("%x", hasher.Sum(nil))

	if result != img.Checksum {
		return fmt.Errorf("[ERROR] Checksum does not match\n Result: %s\n Expected: %s", result, img.Checksum)
	}

	return nil
}
