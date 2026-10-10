package transfer

import (
	"errors"
	"io"
	"testing"
	"time"

	"goexplore/internal/explorer"
)

type holdExplorer struct {
	hold <-chan struct{}
}

func (holdExplorer) Connect() error    { return nil }
func (holdExplorer) Disconnect() error { return nil }
func (holdExplorer) ListDir(string) ([]explorer.FileEntry, error) {
	return nil, nil
}
func (holdExplorer) Stat(string) (explorer.FileEntry, error) {
	return explorer.FileEntry{}, nil
}
func (holdExplorer) MkDir(string) error          { return nil }
func (holdExplorer) Delete(string) error         { return nil }
func (holdExplorer) Rename(string, string) error { return nil }
func (e holdExplorer) ReadFile(string) (io.ReadCloser, error) {
	<-e.hold
	return nil, errors.New("stopped")
}
func (holdExplorer) WriteFile(string, io.Reader, int64) error { return nil }
func (holdExplorer) Checksum(string) (string, error)          { return "", nil }

func TestGetTransfersKeepsQueueOrder(t *testing.T) {
	hold := make(chan struct{})
	src := holdExplorer{hold: hold}
	dst := holdExplorer{hold: hold}
	m := NewManager(1)

	if err := m.QueueTransfer("first", "/a", "/b", "a.txt", 1, src, dst, false, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.QueueTransfer("second", "/c", "/d", "b.txt", 1, src, dst, false, 0); err != nil {
		t.Fatal(err)
	}

	got := transferIDs(m.GetTransfers())
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("order = %v, want [first second]", got)
	}

	close(hold)
	waitTerminal(t, m, "first")
	waitTerminal(t, m, "second")

	m.ClearCompleted()
	if remaining := m.GetTransfers(); len(remaining) != 0 {
		t.Fatalf("cleared transfers = %v, want none", transferIDs(remaining))
	}

	hold = make(chan struct{})
	src.hold = hold
	dst.hold = hold
	if err := m.QueueTransfer("third", "/e", "/f", "c.txt", 1, src, dst, false, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.QueueTransfer("fourth", "/g", "/h", "d.txt", 1, src, dst, false, 0); err != nil {
		t.Fatal(err)
	}

	got = transferIDs(m.GetTransfers())
	if len(got) != 2 || got[0] != "third" || got[1] != "fourth" {
		t.Fatalf("order after clear = %v, want [third fourth]", got)
	}
	close(hold)
}

func transferIDs(transfers []*Transfer) []string {
	ids := make([]string, len(transfers))
	for i, item := range transfers {
		ids[i] = item.ID
	}
	return ids
}

func waitTerminal(t *testing.T, m *Manager, id string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		for _, item := range m.GetTransfers() {
			if item.ID == id && (item.Status == StatusComplete || item.Status == StatusFailed) {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s to finish", id)
}
