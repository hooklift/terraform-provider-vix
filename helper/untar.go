package helper

import (
	"archive/tar"
	"io"
	"os"
)

func Untar(data io.Reader, dest string) error {
	err := os.MkdirAll(dest, 0744)
	if err != nil {
		return err
	}

	tr := tar.NewReader(data)

	// Iterate through the files in the archive.
	returnPath := dest
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
			returnPath += hdr.Name
			os.Mkdir(dest+hdr.Name, 0744)
			continue
		}

		file, err := os.Create(dest + hdr.Name)
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
