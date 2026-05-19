package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var ErrPathTraversal = errors.New("path traversal detected")

// Provider defines the contract for file operations.
type Provider interface {
	Save(ctx context.Context, path string, content io.Reader) error
	Delete(ctx context.Context, path string) error
	AvailableBytes() (int64, error)
}

// LocalStorage implements Provider using the local filesystem.
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	// Ensure base directories exist
	for _, sub := range []string{"uploads", "hls"} {
		if err := os.MkdirAll(filepath.Join(absBase, sub), 0755); err != nil {
			return nil, err
		}
	}
	return &LocalStorage{basePath: absBase}, nil
}

func (l *LocalStorage) isSafePath(path string) (string, error) {
	cleaned := filepath.Clean(path)
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}

	absTarget, err := filepath.Abs(filepath.Join(l.basePath, cleaned))
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(l.basePath, absTarget)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}

	return absTarget, nil
}

func (l *LocalStorage) Save(ctx context.Context, path string, content io.Reader) error {
	fullPath, err := l.isSafePath(path)
	if err != nil {
		return err
	}

	// Ensure parent directory exists (for HLS variants)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, content)
	return err
}

func (l *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath, err := l.isSafePath(path)
	if err != nil {
		return err
	}
	return os.RemoveAll(fullPath)
}

// AvailableBytes returns free space on the storage partition.
func (l *LocalStorage) AvailableBytes() (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(l.basePath, &stat); err != nil {
		return 0, err
	}
	// Available blocks * size per block
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
