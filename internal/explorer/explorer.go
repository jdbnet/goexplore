package explorer

import "io"

type FileEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Modified    string `json:"modified"`
	IsDir       bool   `json:"is_dir"`
	Permissions string `json:"permissions"`
}

type Explorer interface {
	Connect() error
	Disconnect() error
	ListDir(path string) ([]FileEntry, error)
	Stat(path string) (FileEntry, error)
	MkDir(path string) error
	Delete(path string) error
	Rename(src, dst string) error
	ReadFile(path string) (io.ReadCloser, error)
	WriteFile(path string, r io.Reader, size int64) error
}
