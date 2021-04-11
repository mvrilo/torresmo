package torrent

import (
	"encoding/json"
	"sync"

	torren "github.com/anacrolix/torrent"
)

type File interface {
	Priority
	DisplayPath() string
	BytesCompleted() int64
	Length() int64
	Offset() int64
}

type Priority interface {
	GetPriority() byte
	Now()
	High()
	Normal()
	Zero()
}

type file struct {
	*torren.File
	mu  *sync.Mutex
	pri byte
}

func newFile(tf *torren.File, biggestFirst bool) (f *file) {
	f = &file{
		File: tf,
		mu:   new(sync.Mutex),
		pri:  0,
	}

	if biggestFirst {
		f.pri = byte(torren.PiecePriorityNow)
		f.File.SetPriority(torren.PiecePriorityNow)
	}

	return
}

func (f *file) Now() {
	f.pri = byte(torren.PiecePriorityNow)
	f.File.SetPriority(torren.PiecePriorityNow)
}

func (f *file) High() {
	f.pri = byte(torren.PiecePriorityHigh)
	f.File.SetPriority(torren.PiecePriorityHigh)
}

func (f *file) Normal() {
	f.pri = byte(torren.PiecePriorityNormal)
	f.File.SetPriority(torren.PiecePriorityNormal)
}

func (f *file) Zero() {
	f.pri = byte(torren.PiecePriorityNone)
	f.File.SetPriority(torren.PiecePriorityNone)
}

func (f *file) GetPriority() byte {
	return f.pri
}

func (f *file) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		BytesCompleted int64
		DisplayPath    string
		Length         int64
		Offset         int64
		// Priority       byte
	}{
		BytesCompleted: f.BytesCompleted(),
		DisplayPath:    f.DisplayPath(),
		Length:         f.Length(),
		Offset:         f.Offset(),
		// Priority:       f.pri,
	})
}

var _ File = &file{}
