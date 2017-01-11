package validation

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// Checksum calculates a checksum for a file
func Checksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return ChecksumReader(file)
}

// ChecksumReader calculates a checksum for a file
func ChecksumReader(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)[:16]
	return hex.EncodeToString(hashInBytes), nil
}
