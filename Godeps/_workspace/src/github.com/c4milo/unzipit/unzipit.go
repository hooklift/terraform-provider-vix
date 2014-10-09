// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package unzipit

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	magicZIP  []byte = []byte{0x50, 0x4b, 0x03, 0x04}
	magicGZ   []byte = []byte{0x1f, 0x8b}
	magicBZIP []byte = []byte{0x42, 0x5a}
	magicTAR  []byte = []byte{0x75, 0x73, 0x74, 0x61, 0x72} // at offset 257
)

// Check whether a file has the magic number for tar, gzip, bzip2 or zip files
//
// 50 4b 03 04 for pkzip format
// 1f 8b for .gz
// 42 5a for bzip
// 75 73 74 61 72 at offset 257 for tar files
func magicNumber(reader io.ReaderAt, offset int64) (string, error) {
	magic := make([]byte, 5, 5)

	reader.ReadAt(magic, offset)

	if bytes.Equal(magicTAR, magic) {
		return "tar", nil
	}

	if bytes.Equal(magicZIP, magic[0:4]) {
		return "zip", nil
	}

	if bytes.Equal(magicGZ, magic[0:2]) {
		return "gzip", nil
	} else if bytes.Equal(magicBZIP, magic[0:2]) {
		return "bzip", nil
	}

	return "", nil
}

// Unpacks a compressed and archived file and places result output in destination
// path.
//
// File formats supported are:
//   - .tar.gz
//   - .tar.bzip2
//   - .zip
//   - .tar
//
// If it cannot recognize the file format, it will save the file, as is, to the
// destination path.
func Unpack(file *os.File, destPath string) (string, error) {
	if file == nil {
		return "", errors.New("You must provide a valid file to unpack")
	}

	var err error
	if destPath == "" {
		destPath, err = ioutil.TempDir(os.TempDir(), "unpackit-")
		if err != nil {
			return "", err
		}
	}

	// Makes sure despPath exists
	os.MkdirAll(destPath, 0740)

	// Makes sure file cursor is at index 0
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}

	// Reads magic number from file so we can better determine how to proceed
	ftype, err := magicNumber(file, 0)
	if err != nil {
		return "", err
	}

	data := bytes.NewBuffer(nil)
	switch ftype {
	case "gzip":
		data, err = Gunzip(file)
		if err != nil {
			return "", err
		}
	case "bzip":
		data, err = Bunzip2(file)
		if err != nil {
			return "", err
		}
	case "zip":
		// Like TAR, ZIP is also an archiving format, therefore we can just return
		// after it finishes
		return Unzip(file, destPath)
	default:
		io.Copy(data, file)
	}

	// Check magic number in offset 257 too see if this is also a TAR file
	ftype, err = magicNumber(bytes.NewReader(data.Bytes()), 257)
	if ftype == "tar" {
		return Untar(data, destPath)
	}

	// If it's not a TAR archive then save it to disk as is.
	destRawFile := filepath.Join(destPath, sanitize(path.Base(file.Name())))

	// Creates destination file
	destFile, err := os.Create(destRawFile)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	// Copies data to destination file
	if _, err := io.Copy(destFile, data); err != nil {
		return "", err
	}

	return destPath, nil
}

// Decompresses a bzip2 data stream and returns the decompressed stream
func Bunzip2(file *os.File) (*bytes.Buffer, error) {
	data := bzip2.NewReader(file)

	buffer := bytes.NewBuffer(nil)
	io.Copy(buffer, data)

	return buffer, nil
}

// Decompresses a gzip data stream and returns the decompressed stream
func Gunzip(file *os.File) (*bytes.Buffer, error) {
	data, err := gzip.NewReader(file)
	if err != nil && err != io.EOF {
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	io.Copy(buffer, data)

	return buffer, nil
}

// Decompresses and unarchives a ZIP archive, returning the final path or an error
func Unzip(file *os.File, destPath string) (string, error) {
	// Open a zip archive for reading.
	r, err := zip.OpenReader(file.Name())
	if err != nil {
		return "", err
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		file, err := os.Create(filepath.Join(destPath, sanitize(f.Name)))
		if err != nil {
			return "", err
		}
		defer file.Close()

		if _, err := io.CopyN(file, rc, int64(f.UncompressedSize64)); err != nil {
			return "", err
		}
		defer rc.Close()
	}
	return destPath, nil
}

// Unarchives a TAR archive and returns the final destination path or an error
func Untar(data io.Reader, destPath string) (string, error) {
	// Makes sure destPath exists
	os.MkdirAll(destPath, 0740)

	tr := tar.NewReader(data)

	// Iterate through the files in the archive.
	rootdir := destPath
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
			d := filepath.Join(destPath, sanitize(hdr.Name))
			if rootdir == destPath {
				rootdir = d
			}
			os.Mkdir(d, 0740)
			continue
		}

		file, err := os.Create(filepath.Join(destPath, sanitize(hdr.Name)))
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

// Sanitizes name to avoid overwriting sensitive system files when unarchiving
func sanitize(name string) string {
	// Gets rid of volume drive label in Windows
	if len(name) > 1 && name[1] == ':' && runtime.GOOS == "windows" {
		name = name[2:]
	}

	name = filepath.Clean(name)
	name = filepath.ToSlash(name)
	for strings.HasPrefix(name, "../") {
		name = name[3:]
	}
	return name
}
