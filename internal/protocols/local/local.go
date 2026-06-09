package local

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"goexplore/internal/explorer"
)

type LocalExplorer struct {
	basePath string
}

func New() *LocalExplorer {
	basePath := "/"
	if runtime.GOOS == "windows" {
		basePath = "C:\\"
	}
	return &LocalExplorer{basePath: basePath}
}

func (e *LocalExplorer) Connect() error    { return nil }
func (e *LocalExplorer) Disconnect() error { return nil }

func (e *LocalExplorer) ListDir(path string) ([]explorer.FileEntry, error) {
	if path == "" {
		path = e.basePath
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []explorer.FileEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, explorer.FileEntry{
			Name:        entry.Name(),
			Path:        filepath.ToSlash(filepath.Join(path, entry.Name())),
			Size:        info.Size(),
			Modified:    info.ModTime().Format(time.RFC3339),
			IsDir:       entry.IsDir(),
			Permissions: info.Mode().String(),
		})
	}
	return result, nil
}

func (e *LocalExplorer) Stat(path string) (explorer.FileEntry, error) {
	if path == "" {
		path = e.basePath
	}
	info, err := os.Stat(path)
	if err != nil {
		return explorer.FileEntry{}, err
	}
	return explorer.FileEntry{
		Name:        info.Name(),
		Path:        filepath.ToSlash(path),
		Size:        info.Size(),
		Modified:    info.ModTime().Format(time.RFC3339),
		IsDir:       info.IsDir(),
		Permissions: info.Mode().String(),
	}, nil
}

func (e *LocalExplorer) MkDir(path string) error {
	return os.Mkdir(path, 0755)
}

func (e *LocalExplorer) Delete(path string) error {
	return os.RemoveAll(path)
}

func (e *LocalExplorer) Rename(src, dst string) error {
	return os.Rename(src, dst)
}

func (e *LocalExplorer) Checksum(path string) (string, error) {
	r, err := e.ReadFile(path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (e *LocalExplorer) ReadFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (e *LocalExplorer) WriteFile(path string, r io.Reader, size int64) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
