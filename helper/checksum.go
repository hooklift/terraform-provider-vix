package helper

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"log"
)

func VerifyChecksum(data io.Reader, algorithm, sum string) error {
	log.Printf("[DEBUG] Verifying checksum...")
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
	_, err := io.Copy(hasher, data)
	if err != nil {
		return err
	}

	result := fmt.Sprintf("%x", hasher.Sum(nil))

	if result != sum {
		return fmt.Errorf("Checksum does not match\n Result: %s\n Expected: %s", result, sum)
	}

	return nil
}
