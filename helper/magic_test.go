package helper

import (
	"os"
	"testing"
)

func TestMagicNumber(t *testing.T) {
	var tests = []struct {
		filepath string
		offset   int64
		ftype    string
	}{
		{"./fixtures/test.tar.bzip2", 0, "bzip"},
		{"./fixtures/test.tar.gz", 0, "gz"},
		{"./fixtures/test.zip", 0, "zip"},
		{"./fixtures/test.tar", 257, "tar"},
	}

	for _, test := range tests {
		file, err := os.Open(test.filepath)
		ok(t, err)
		defer file.Close()
		ftype, err := MagicNumber(file, test.offset)
		ok(t, err)

		assert(t, ftype == test.ftype, ftype+" != "+test.ftype)
	}
}
