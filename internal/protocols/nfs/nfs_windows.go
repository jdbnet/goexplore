//go:build windows

package nfs

import (
	"fmt"
	"io"

	"goexplore/internal/config"
	"goexplore/internal/explorer"
)

type NFSExplorer struct{}

func New(c *config.ConnectionConfig, secret string) (*NFSExplorer, error) {
	return nil, fmt.Errorf("NFS is not supported on Windows")
}

func (e *NFSExplorer) Connect() error { return nil }
func (e *NFSExplorer) Disconnect() error { return nil }
func (e *NFSExplorer) ListDir(path string) ([]explorer.FileEntry, error) { return nil, nil }
func (e *NFSExplorer) Stat(path string) (explorer.FileEntry, error) { return explorer.FileEntry{}, nil }
func (e *NFSExplorer) MkDir(path string) error { return nil }
func (e *NFSExplorer) Delete(path string) error { return nil }
func (e *NFSExplorer) Rename(src, dst string) error { return nil }
func (e *NFSExplorer) ReadFile(path string) (io.ReadCloser, error) { return nil, nil }
func (e *NFSExplorer) WriteFile(path string, r io.Reader, size int64) error { return nil }
func (e *NFSExplorer) Checksum(path string) (string, error) { return "", nil }
