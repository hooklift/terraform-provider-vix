package helper_test

import (
	"testing"
	"github.com/c4milo/terraform_vix/helper"
)

func TestVerifyChecksum(t *testing.T) {
	//echo -n Test | gsha256sum
	ok(t, helper.VerifyChecksum([]byte("Test"), "sha256", "532eaabd9574880dbf76b9b8cc00832c20a6ec113d682299550d7a6e0f345e25"))
	//echo -n Test | gsha1sum
	ok(t, helper.VerifyChecksum([]byte("Test"), "sha1", "640ab2bae07bedc4c163f679a746f7ab7fb5d1fa"))
	//echo -n Test | gsha512sum
	ok(t, helper.VerifyChecksum([]byte("Test"), "sha512", "c6ee9e33cf5c6715a1d148fd73f7318884b41adcb916021e2bc0e800a5c5dd97f5142178f6ae88c8fdd98e1afb0ce4c8d2c54b5f37b30b7da1997bb33b0b8a31"))
	//echo -n Test | gmd5sum
	ok(t, helper.VerifyChecksum([]byte("Test"), "md5", "0cbc6611f5540bd0809a388dc95a615b"))
}
