// Package util contains a couple of utility methods.
package util

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// HashFile returns the SHA1-hash of the contents of the specified file.
func HashFile(filePath string) (string, error) {
	var returnSHA1String string

	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}

	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}

	hashInBytes := hash.Sum(nil)[:20]
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil
}

// HasSSHAgent reports whether the SSH agent is available
func HasSSHAgent() bool {
	authsock, ok := os.LookupEnv("SSH_AUTH_SOCK")
	if !ok {
		return false
	}
	if dirent, err := os.Stat(authsock); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		if dirent.Mode()&os.ModeSocket == 0 {
			return false
		}
	}
	return true
}
