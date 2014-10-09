// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package unzipit

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func TestUnpack(t *testing.T) {
	var tests = []struct {
		filepath string
		files    int
	}{
		{"./fixtures/test.tar.bzip2", 2},
		{"./fixtures/test.tar.gz", 2},
		{"./fixtures/test.zip", 2},
		{"./fixtures/test.tar", 2},
		{"./fixtures/cfgdrv.iso", 1},
		{"./fixtures/test2.tar.gz", 4},
	}

	for _, test := range tests {
		tempDir, err := ioutil.TempDir(os.TempDir(), "unpackit-tests-"+path.Base(test.filepath)+"-")
		ok(t, err)
		defer os.RemoveAll(tempDir)

		file, err := os.Open(test.filepath)
		ok(t, err)
		defer file.Close()

		destPath, err := Unpack(file, tempDir)
		ok(t, err)

		finfo, err := ioutil.ReadDir(destPath)
		ok(t, err)

		length := len(finfo)
		assert(t, length == test.files, fmt.Sprintf("%d != %d for %s", length, test.files, destPath))
	}
}

func TestMagicNumber(t *testing.T) {
	var tests = []struct {
		filepath string
		offset   int64
		ftype    string
	}{
		{"./fixtures/test.tar.bzip2", 0, "bzip"},
		{"./fixtures/test.tar.gz", 0, "gzip"},
		{"./fixtures/test.zip", 0, "zip"},
		{"./fixtures/test.tar", 257, "tar"},
	}

	for _, test := range tests {
		file, err := os.Open(test.filepath)
		ok(t, err)

		ftype, err := magicNumber(file, test.offset)
		file.Close()
		ok(t, err)

		assert(t, ftype == test.ftype, ftype+" != "+test.ftype)
	}
}

func TestUntar(t *testing.T) {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new tar archive.
	tw := tar.NewWriter(buf)

	// Add some files to the archive.
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling licence."},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Size: int64(len(file.Body)),
		}
		err := tw.WriteHeader(hdr)
		ok(t, err)

		_, err = tw.Write([]byte(file.Body))
		ok(t, err)
	}

	// Make sure to check the error on Close.
	err := tw.Close()
	ok(t, err)

	// Open the tar archive for reading.
	r := bytes.NewReader(buf.Bytes())
	destDir, err := ioutil.TempDir(os.TempDir(), "terraform-vix")
	ok(t, err)
	defer os.RemoveAll(destDir)

	_, err = Untar(r, destDir)
	ok(t, err)
}

func TestSanitize(t *testing.T) {
	var tests = []struct {
		malicious string
		sanitized string
	}{
		{"../../.././etc/passwd", "etc/passwd"},
		{"../../etc/passwd", "etc/passwd"},
		{"./etc/passwd", "etc/passwd"},
		{"./././etc/passwd", "etc/passwd"},
		{"nonexistant/b/../file.txt", "nonexistant/file.txt"},
		{"abc../def", "abc../def"},
		{"a/b/c/../d", "a/b/d"},
		{"a/../../c", "c"},
		{"...../etc/password", "...../etc/password"},
	}

	for _, test := range tests {
		a := sanitize(test.malicious)
		msg := fmt.Sprintf("%s != %s for malicious string %s", a, test.sanitized, test.malicious)
		assert(t, a == test.sanitized, msg)
	}
}
