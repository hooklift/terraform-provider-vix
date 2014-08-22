package helper_test

import (
	"bytes"
	"testing"

	"github.com/c4milo/terraform-provider-vix/helper"
)

// For some unknown reason if we take bytes.NewBuffer([]byte("Test"))
// to a variable, only the first test passes. The rest of the tests fail.
// There is some state being held by Go or some undocumented condition
// that we should be aware of.
func TestVerifyChecksum(t *testing.T) {
	data := bytes.NewBuffer([]byte("Test"))
	var tests = []struct {
		data      *bytes.Buffer
		algorithm string
		checksum  string
	}{
		// echo -n Test | gsha256sum
		{data, "sha256", "532eaabd9574880dbf76b9b8cc00832c20a6ec113d682299550d7a6e0f345e25"},
		// echo -n Test | gsha1sum
		{data, "sha1", "640ab2bae07bedc4c163f679a746f7ab7fb5d1fa"},
		// echo -n Test | gsha512sum
		{data, "sha512", "c6ee9e33cf5c6715a1d148fd73f7318884b41adcb916021e2bc0e800a5c5dd97f5142178f6ae88c8fdd98e1afb0ce4c8d2c54b5f37b30b7da1997bb33b0b8a31"},
		// echo -n Test | gmd5sum
		{data, "md5", "0cbc6611f5540bd0809a388dc95a615b"},
	}

	for _, test := range tests {
		buf := bytes.NewReader(test.data.Bytes())
		ok(t, helper.VerifyChecksum(buf, test.algorithm, test.checksum))
	}
}
