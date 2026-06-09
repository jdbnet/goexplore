package transfer

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"goexplore/internal/explorer"
)

type Status string

const (
	StatusQueued   Status = "queued"
	StatusActive   Status = "active"
	StatusComplete Status = "complete"
	StatusFailed   Status = "failed"
)

type Transfer struct {
	ID          string  `json:"id"`
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Filename    string  `json:"filename"`
	BytesTotal  int64   `json:"bytes_total"`
	BytesDone   int64   `json:"bytes_done"`
	SpeedMBps   float64 `json:"speed_mbps"`
	ETA         int     `json:"eta_seconds"`
	Status      Status  `json:"status"`
	Error       string  `json:"error,omitempty"`
	Verify      bool    `json:"verify"`
	LimitMBps   int     `json:"limit_mbps"`

	srcExp explorer.Explorer
	dstExp explorer.Explorer

	startTime time.Time
}

type Manager struct {
	mu          sync.Mutex
	transfers   map[string]*Transfer
	queue       chan *Transfer
	concurrency int
}

func NewManager(concurrency int) *Manager {
	m := &Manager{
		transfers:   make(map[string]*Transfer),
		queue:       make(chan *Transfer, 100),
		concurrency: concurrency,
	}
	for i := 0; i < concurrency; i++ {
		go m.worker()
	}
	return m
}

func (m *Manager) worker() {
	for t := range m.queue {
		m.process(t)
	}
}

func (m *Manager) process(t *Transfer) {
	t.Status = StatusActive
	t.startTime = time.Now()
	err := m.doTransfer(t)
	if err != nil {
		t.Status = StatusFailed
		t.Error = err.Error()
	} else {
		t.Status = StatusComplete
		t.BytesDone = t.BytesTotal
	}
}

func (m *Manager) doTransfer(t *Transfer) error {
	r, err := t.srcExp.ReadFile(t.Source)
	if err != nil {
		return err
	}
	defer r.Close()

	hash := md5.New()
	tr := io.TeeReader(r, hash)

	var streamReader io.Reader = tr
	if t.LimitMBps > 0 {
		limitBytes := t.LimitMBps * 1024 * 1024
		// Set burst to limitBytes to allow typical io.Copy chunk sizes without erroring
		limiter := rate.NewLimiter(rate.Limit(limitBytes), limitBytes)
		streamReader = &throttledReader{
			r:       tr,
			limiter: limiter,
		}
	}

	pr := &progressReader{
		r: streamReader,
		t: t,
	}

	if err := t.dstExp.WriteFile(t.Destination, pr, t.BytesTotal); err != nil {
		return err
	}

	if t.Verify {
		expectedHash := hex.EncodeToString(hash.Sum(nil))
		t.Status = StatusActive // Still active during verification

		actualHash, err := t.dstExp.Checksum(t.Destination)
		if err != nil {
			return fmt.Errorf("verification failed: could not calculate destination checksum: %w", err)
		}

		if expectedHash != actualHash {
			// S3 Multi-part ETags end with -N, we can't easily verify them with local md5
			if !strings.Contains(actualHash, "-") {
				return fmt.Errorf("verification failed: checksum mismatch (expected %s, got %s)", expectedHash, actualHash)
			}
		}
	}

	return nil
}

type throttledReader struct {
	r       io.Reader
	limiter *rate.Limiter
}

func (tr *throttledReader) Read(p []byte) (int, error) {
	n, err := tr.r.Read(p)
	if n > 0 && tr.limiter != nil {
		if errWait := tr.limiter.WaitN(context.Background(), n); errWait != nil {
			return n, errWait
		}
	}
	return n, err
}

type progressReader struct {
	r io.Reader
	t *Transfer
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.t.BytesDone += int64(n)
		elapsed := time.Since(pr.t.startTime).Seconds()
		if elapsed > 0 {
			pr.t.SpeedMBps = (float64(pr.t.BytesDone) / 1024 / 1024) / elapsed
			if pr.t.SpeedMBps > 0 {
				pr.t.ETA = int(float64(pr.t.BytesTotal-pr.t.BytesDone) / 1024 / 1024 / pr.t.SpeedMBps)
			}
		}
	}
	return n, err
}

func (m *Manager) QueueTransfer(id, srcPath, dstPath, filename string, size int64, srcExp, dstExp explorer.Explorer, verify bool, limitMBps int) error {
	if srcExp == nil || dstExp == nil {
		return errors.New("invalid explorers")
	}

	t := &Transfer{
		ID:          id,
		Source:      srcPath,
		Destination: dstPath,
		Filename:    filename,
		BytesTotal:  size,
		Status:      StatusQueued,
		Verify:      verify,
		LimitMBps:   limitMBps,
		srcExp:      srcExp,
		dstExp:      dstExp,
	}

	m.mu.Lock()
	m.transfers[id] = t
	m.mu.Unlock()

	m.queue <- t
	return nil
}

func (m *Manager) GetTransfers() []*Transfer {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]*Transfer, 0, len(m.transfers))
	for _, t := range m.transfers {
		res = append(res, t)
	}
	return res
}

func (m *Manager) ClearCompleted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, t := range m.transfers {
		if t.Status == StatusComplete || t.Status == StatusFailed {
			delete(m.transfers, id)
		}
	}
}
