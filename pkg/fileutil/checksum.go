package fileutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileChecksum calculates SHA256 checksum of a file
func FileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

// VerifyChecksum verifies a file's checksum
func VerifyChecksum(path string, expectedChecksum string) (bool, error) {
	actualChecksum, err := FileChecksum(path)
	if err != nil {
		return false, err
	}
	return actualChecksum == expectedChecksum, nil
}

// DirChecksum calculates a combined checksum for all files in a directory
func DirChecksum(dir string) (string, error) {
	hash := sha256.New()

	err := WalkFiles(dir, func(path string, info os.FileInfo) error {
		// Add the relative path to the hash
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		hash.Write([]byte(relPath))

		// Add file content hash
		fileHash, err := FileChecksum(path)
		if err != nil {
			return err
		}
		hash.Write([]byte(fileHash))

		return nil
	})

	if err != nil {
		return "", err
	}

	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

// WalkFiles walks through all files in a directory (not directories themselves)
func WalkFiles(root string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fn(path, info)
		}
		return nil
	})
}
