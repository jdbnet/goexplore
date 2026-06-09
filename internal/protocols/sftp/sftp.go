package sftp

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"goexplore/internal/config"
	"goexplore/internal/explorer"
)

type SFTPExplorer struct {
	cfg    *config.ConnectionConfig
	secret string
	client *sftp.Client
	conn   *ssh.Client
}

func New(c *config.ConnectionConfig, secret string) *SFTPExplorer {
	return &SFTPExplorer{cfg: c, secret: secret}
}

func (e *SFTPExplorer) Connect() error {
	port := e.cfg.Port
	if port == 0 {
		port = 22
	}
	
	var authMethod ssh.AuthMethod
	signer, err := ssh.ParsePrivateKey([]byte(e.secret))
	if err == nil {
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(e.secret)
	}

	config := &ssh.ClientConfig{
		User: e.cfg.Username,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", e.cfg.Host, port), config)
	if err != nil {
		return err
	}
	e.conn = conn

	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return err
	}
	e.client = client
	return nil
}

func (e *SFTPExplorer) Disconnect() error {
	if e.client != nil {
		e.client.Close()
	}
	if e.conn != nil {
		e.conn.Close()
	}
	return nil
}

func (e *SFTPExplorer) ListDir(path string) ([]explorer.FileEntry, error) {
	if path == "" {
		path = "."
	}
	files, err := e.client.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var entries []explorer.FileEntry
	for _, f := range files {
		entries = append(entries, explorer.FileEntry{
			Name:        f.Name(),
			Path:        filepath.ToSlash(filepath.Join(path, f.Name())),
			Size:        f.Size(),
			Modified:    f.ModTime().Format(time.RFC3339),
			IsDir:       f.IsDir(),
			Permissions: f.Mode().String(),
		})
	}
	return entries, nil
}

func (e *SFTPExplorer) Stat(path string) (explorer.FileEntry, error) {
	if path == "" {
		path = "."
	}
	f, err := e.client.Stat(path)
	if err != nil {
		return explorer.FileEntry{}, err
	}
	return explorer.FileEntry{
		Name:        f.Name(),
		Path:        filepath.ToSlash(path),
		Size:        f.Size(),
		Modified:    f.ModTime().Format(time.RFC3339),
		IsDir:       f.IsDir(),
		Permissions: f.Mode().String(),
	}, nil
}

func (e *SFTPExplorer) MkDir(path string) error {
	return e.client.MkdirAll(path)
}

func (e *SFTPExplorer) removeAll(path string) error {
	fi, err := e.client.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return e.client.Remove(path)
	}

	files, err := e.client.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.Name() == "." || f.Name() == ".." {
			continue
		}
		if err := e.removeAll(filepath.Join(path, f.Name())); err != nil {
			return err
		}
	}

	return e.client.RemoveDirectory(path)
}

func (e *SFTPExplorer) Delete(path string) error {
	return e.removeAll(path)
}

func (e *SFTPExplorer) Rename(src, dst string) error {
	return e.client.Rename(src, dst)
}

func (e *SFTPExplorer) ReadFile(path string) (io.ReadCloser, error) {
	return e.client.Open(path)
}

func (e *SFTPExplorer) WriteFile(path string, r io.Reader, size int64) error {
	f, err := e.client.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
