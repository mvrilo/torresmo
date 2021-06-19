package torrent

import (
	"path/filepath"
	"sync"

	torren "github.com/anacrolix/torrent"
)

type File interface {
	Priority
	DisplayPath() string
	Path() string
	BytesCompleted() int64
	Length() int64
	Offset() int64
	Name() string
}

type Priority interface {
	GetPriority() byte
	Now()
	High()
	Normal()
	Zero()
}

var _ File = (*file)(nil)

type file struct {
	*torren.File
	mu  sync.Mutex
	pri byte
}

func newFile(tf *torren.File) (f *file) {
	f = &file{
		File: tf,
		mu:   sync.Mutex{},
		pri:  0,
	}

	return
}

func (f *file) Name() string {
	_, name := filepath.Split(f.File.Path())
	return name
}

func (f *file) Now() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pri = byte(torren.PiecePriorityNow)
	f.File.SetPriority(torren.PiecePriorityNow)
}

func (f *file) High() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pri = byte(torren.PiecePriorityHigh)
	f.File.SetPriority(torren.PiecePriorityHigh)
}

func (f *file) Normal() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pri = byte(torren.PiecePriorityNormal)
	f.File.SetPriority(torren.PiecePriorityNormal)
}

func (f *file) Zero() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pri = byte(torren.PiecePriorityNone)
	f.File.SetPriority(torren.PiecePriorityNone)
}

func (f *file) GetPriority() byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.pri
}

func (f *file) MarshalJSON() ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return json.Marshal(struct {
		Completed      bool   `json:"completed"`
		BytesCompleted int64  `json:"bytesCompleted"`
		Length         int64  `json:"length"`
		Offset         int64  `json:"offset"`
		Path           string `json:"path"`
		DisplayPath    string `json:"displayPath"`
		Name           string `json:"name"`
	}{
		Completed:      f.BytesCompleted() == f.Length(),
		BytesCompleted: f.BytesCompleted(),
		Path:           f.Path(),
		DisplayPath:    f.DisplayPath(),
		Length:         f.Length(),
		Offset:         f.Offset(),
		Name:           f.Name(),
	})
}
