package main

import (
	"context"
	"fmt"
	"goexplore/internal/config"
	"goexplore/internal/explorer"
	"goexplore/internal/keychain"
	"goexplore/internal/protocols/ftp"
	"goexplore/internal/protocols/local"
	"goexplore/internal/protocols/nfs"
	"goexplore/internal/protocols/s3"
	"goexplore/internal/protocols/sftp"
	"goexplore/internal/protocols/smb"
	"goexplore/internal/protocols/webdav"
	"goexplore/internal/transfer"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx             context.Context
	cfg             *config.Config
	transferManager *transfer.Manager
}

func NewApp() *App {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
	}
	if cfg == nil {
		cfg = &config.Config{Connections: []config.ConnectionConfig{}}
	}
	return &App{
		cfg:             cfg,
		transferManager: transfer.NewManager(3),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetVersion() string {
	return Version
}

func (a *App) GetConnections() []config.ConnectionConfig {
	if a.cfg == nil {
		return nil
	}
	return a.cfg.Connections
}

func (a *App) SaveConnection(c config.ConnectionConfig, secret string) error {
	if c.ID == "" {
		return fmt.Errorf("connection ID cannot be empty")
	}

	if secret != "" {
		err := keychain.SetSecret(c.ID, secret)
		if err != nil {
			return err
		}
	}

	found := false
	for i, existing := range a.cfg.Connections {
		if existing.ID == c.ID {
			a.cfg.Connections[i] = c
			found = true
			break
		}
	}
	if !found {
		a.cfg.Connections = append(a.cfg.Connections, c)
	}

	return config.SaveConfig(a.cfg)
}

func (a *App) DeleteConnection(id string) error {
	newConns := []config.ConnectionConfig{}
	for _, c := range a.cfg.Connections {
		if c.ID != id {
			newConns = append(newConns, c)
		}
	}
	a.cfg.Connections = newConns
	_ = keychain.DeleteSecret(id)
	return config.SaveConfig(a.cfg)
}

func (a *App) getExplorerForConnection(id string) (explorer.Explorer, error) {
	if id == "local" {
		return local.New(), nil
	}

	var conn *config.ConnectionConfig
	for _, c := range a.cfg.Connections {
		if c.ID == id {
			conn = &c
			break
		}
	}
	if conn == nil {
		return nil, fmt.Errorf("connection not found")
	}

	secret, _ := keychain.GetSecret(id)

	switch conn.Protocol {
	case "s3":
		return s3.New(conn, secret), nil
	case "sftp":
		return sftp.New(conn, secret), nil
	case "smb":
		return smb.New(conn, secret), nil
	case "webdav":
		return webdav.New(conn, secret), nil
	case "nfs":
		return nfs.New(conn, secret)
	case "ftp":
		return ftp.New(conn, secret), nil
	default:
		return nil, fmt.Errorf("protocol %s not fully implemented", conn.Protocol)
	}
}

func (a *App) ListDir(connId, path string) ([]explorer.FileEntry, error) {
	exp, err := a.getExplorerForConnection(connId)
	if err != nil {
		return nil, err
	}
	if err := exp.Connect(); err != nil {
		return nil, err
	}
	defer exp.Disconnect()
	return exp.ListDir(path)
}

func (a *App) MkDir(connId, path string) error {
	exp, err := a.getExplorerForConnection(connId)
	if err != nil {
		return err
	}
	if err := exp.Connect(); err != nil {
		return err
	}
	defer exp.Disconnect()
	return exp.MkDir(path)
}

func (a *App) Delete(connId, path string) error {
	exp, err := a.getExplorerForConnection(connId)
	if err != nil {
		return err
	}
	if err := exp.Connect(); err != nil {
		return err
	}
	defer exp.Disconnect()
	return exp.Delete(path)
}

func (a *App) QueueTransfer(id string, srcConnId, dstConnId, srcPath, dstPath, filename string, size int64, verify bool) error {
	srcExp, err := a.getExplorerForConnection(srcConnId)
	if err != nil {
		return err
	}
	dstExp, err := a.getExplorerForConnection(dstConnId)
	if err != nil {
		return err
	}

	if err := srcExp.Connect(); err != nil {
		return err
	}
	if err := dstExp.Connect(); err != nil {
		srcExp.Disconnect()
		return err
	}

	return a.transferManager.QueueTransfer(id, srcPath, dstPath, filename, size, srcExp, dstExp, verify)
}

func (a *App) GetTransfers() []*transfer.Transfer {
	return a.transferManager.GetTransfers()
}

func (a *App) ClearTransfers() {
	a.transferManager.ClearCompleted()
}

func (a *App) Rename(connId, src, dst string) error {
	exp, err := a.getExplorerForConnection(connId)
	if err != nil {
		return err
	}
	if err := exp.Connect(); err != nil {
		return err
	}
	defer exp.Disconnect()
	return exp.Rename(src, dst)
}

func (a *App) PromptUploadFiles(connId, destPath string) error {
	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Files to Upload",
	})
	if err != nil {
		return err
	}

	for _, localPath := range files {
		stat, err := os.Stat(localPath)
		if err != nil {
			continue
		}
		id := uuid.New().String()
		fileName := filepath.Base(localPath)
		remotePath := destPath
		if destPath == "" || destPath == "/" {
			remotePath = fileName
		} else if destPath[len(destPath)-1] != '/' {
			remotePath = destPath + "/" + fileName
		} else {
			remotePath = destPath + fileName
		}

		a.QueueTransfer(id, "local", connId, localPath, remotePath, fileName, stat.Size(), false)
	}
	return nil
}

func (a *App) PromptUploadDirectory(connId, destPath string) error {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory to Upload",
	})
	if err != nil || dir == "" {
		return err
	}

	baseDirName := filepath.Base(dir)
	remoteBaseDir := destPath
	if remoteBaseDir == "" || remoteBaseDir == "/" {
		remoteBaseDir = baseDirName
	} else if remoteBaseDir[len(remoteBaseDir)-1] != '/' {
		remoteBaseDir = remoteBaseDir + "/" + baseDirName
	} else {
		remoteBaseDir = remoteBaseDir + baseDirName
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(dir, path)
		if relPath == "." {
			return a.MkDir(connId, remoteBaseDir)
		}

		remotePath := remoteBaseDir + "/" + filepath.ToSlash(relPath)

		if info.IsDir() {
			return a.MkDir(connId, remotePath)
		}

		id := uuid.New().String()
		a.QueueTransfer(id, "local", connId, path, remotePath, info.Name(), info.Size(), false)
		return nil
	})
}

func (a *App) PromptDownload(connId, remotePath string) error {
	fileName := filepath.Base(remotePath)
	localPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save File",
		DefaultFilename: fileName,
	})
	if err != nil || localPath == "" {
		return err
	}

	id := uuid.New().String()
	
	exp, err := a.getExplorerForConnection(connId)
	if err != nil {
		return err
	}
	if err := exp.Connect(); err != nil {
		return err
	}
	stat, err := exp.Stat(remotePath)
	exp.Disconnect()
	
	size := int64(0)
	if err == nil {
		size = stat.Size
	}

	return a.QueueTransfer(id, connId, "local", remotePath, localPath, fileName, size, false)
}

type TransferItem struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

func (a *App) TransferItems(srcConnId, dstConnId, dstPath string, items []TransferItem, verify bool) error {
	for _, item := range items {
		if !item.IsDir {
			id := uuid.New().String()
			
			remotePath := dstPath
			if dstPath == "" || dstPath == "/" {
				remotePath = item.Name
			} else if dstPath[len(dstPath)-1] != '/' {
				remotePath = dstPath + "/" + item.Name
			} else {
				remotePath = dstPath + item.Name
			}

			if err := a.QueueTransfer(id, srcConnId, dstConnId, item.Path, remotePath, item.Name, item.Size, verify); err != nil {
				return err
			}
		} else {
			// Run remote walk in goroutine to not block UI
			go a.transferRemoteDirectory(srcConnId, dstConnId, item.Path, dstPath, verify)
		}
	}
	return nil
}

func (a *App) transferRemoteDirectory(srcConnId, dstConnId, srcDirPath, dstBasePath string, verify bool) {
	baseName := filepath.Base(srcDirPath)
	
	newDstPath := dstBasePath
	if dstBasePath == "" || dstBasePath == "/" {
		newDstPath = baseName
	} else if dstBasePath[len(dstBasePath)-1] != '/' {
		newDstPath = dstBasePath + "/" + baseName
	} else {
		newDstPath = dstBasePath + baseName
	}

	a.MkDir(dstConnId, newDstPath)

	entries, err := a.ListDir(srcConnId, srcDirPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir {
			a.transferRemoteDirectory(srcConnId, dstConnId, entry.Path, newDstPath, verify)
		} else {
			id := uuid.New().String()
			itemDstPath := newDstPath + "/" + entry.Name
			a.QueueTransfer(id, srcConnId, dstConnId, entry.Path, itemDstPath, entry.Name, entry.Size, verify)
		}
	}
}
