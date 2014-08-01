package helper

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

func Untar(data io.Reader, dest string) error {
	os.MkdirAll(dest, 0740)

	tr := tar.NewReader(data)

	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}

		if err != nil {
			return err
		}

		if hdr.FileInfo().IsDir() {
			os.Mkdir(filepath.Join(dest, hdr.Name), 0740)
			continue
		}

		file, err := os.Create(filepath.Join(dest, hdr.Name))
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(file, tr); err != nil {
			return err
		}
	}

	return nil
}
