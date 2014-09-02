package helper

import (
	"bytes"
	"os"
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

func MagicNumber(file *os.File, offset int64) (string, error) {
	magic := make([]byte, 5, 5)

	file.ReadAt(magic, offset)

	if bytes.Equal(magicTAR, magic) {
		return "tar", nil
	}

	if bytes.Equal(magicZIP, magic[0:4]) {
		return "zip", nil
	}

	if bytes.Equal(magicGZ, magic[0:2]) {
		return "gz", nil
	} else if bytes.Equal(magicBZIP, magic[0:2]) {
		return "bzip", nil
	}

	return "", nil
}
