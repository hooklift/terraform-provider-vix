package helper

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

func VerifyChecksum(data []byte, algorithm, sum string) error {
	var hasher hash.Hash

	switch algorithm {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		return fmt.Errorf("Crypto algorithm no supported: %s", algorithm)
	}

	_, err := io.Copy(hasher, bytes.NewReader(data))
	if err != nil {
		return err
	}

	result := hex.EncodeToString(hasher.Sum(nil))

	if result != sum {
		return fmt.Errorf("Checksum does not match\n Result: %s\n Expected: %s", result, sum)
	}

	return nil
}
