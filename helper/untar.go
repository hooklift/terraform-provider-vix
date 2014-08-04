package helper

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Untar(data io.Reader, dest string) (string, error) {
	log.Printf("[DEBUG] Untaring archive into %s ...", dest)

	os.MkdirAll(dest, 0740)

	tr := tar.NewReader(data)

	// Iterate through the files in the archive.
	rootdir := dest
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}

		if err != nil {
			return rootdir, err
		}

		if hdr.FileInfo().IsDir() {
			d := filepath.Join(dest, hdr.Name)
			if rootdir == dest {
				rootdir = d
			}
			os.Mkdir(d, 0740)
			continue
		}

		file, err := os.Create(filepath.Join(dest, hdr.Name))
		if err != nil {
			return rootdir, err
		}
		defer file.Close()

		if _, err := io.Copy(file, tr); err != nil {
			return rootdir, err
		}
	}

	return rootdir, nil
}
