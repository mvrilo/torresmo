package torrent

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	torren "github.com/anacrolix/torrent"
)

type Torrent interface {
	Files() []File
	InfoHash() []byte
	Name() string
	TotalLength() uint64
	URI() string
	BytesCompleted() uint64
	Completed() bool
	MarshalJSON() ([]byte, error)
	String() string
}

type entry struct {
	uri            string
	name           string
	numPieces      int
	files          []File
	nodes          []string
	infoHash       []byte
	totalLength    uint64
	bytesCompleted uint64
	pieceLength    uint64
}

func biggestFile(t Torrent) (biggest File) {
	for _, f := range t.Files() {
		if biggest == nil {
			biggest = f
			continue
		}

		if f.Length() > biggest.Length() {
			biggest = f
		}
	}
	return
}

func newTorrent(t *torren.Torrent, biggestFirst bool) Torrent {
	var name string
	var numPieces int
	var totalLength uint64
	var files []File
	var nodes []string
	var pieceLength uint64

	if info := t.Info(); info != nil {
		name = info.Name
		numPieces = info.NumPieces()
		totalLength = uint64(info.TotalLength())
		pieceLength = uint64(info.PieceLength)

		for _, f := range t.Files() {
			files = append(files, newFile(f, biggestFirst))
		}
	}

	metainfo := t.Metainfo()
	for _, node := range metainfo.Nodes {
		nodes = append(nodes, string(node))
	}

	return entry{
		files:          files,
		nodes:          nodes,
		name:           name,
		numPieces:      numPieces,
		totalLength:    totalLength,
		infoHash:       metainfo.HashInfoBytes().Bytes(),
		bytesCompleted: uint64(t.BytesCompleted()),
		pieceLength:    pieceLength,
	}
}

func (e entry) Files() []File          { return e.files }
func (e entry) Downloaded() int        { return e.numPieces }
func (e entry) InfoHash() []byte       { return e.infoHash }
func (e entry) Name() string           { return e.name }
func (e entry) URI() string            { return e.uri }
func (e entry) TotalLength() uint64    { return e.totalLength }
func (e entry) BytesCompleted() uint64 { return e.bytesCompleted }
func (e entry) Completed() bool        { return e.BytesCompleted() == e.TotalLength() }

func (e entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		// Nodes []
		Files          []File `json:"files"`
		Name           string `json:"name"`
		NumPieces      int    `json:"numPieces"`
		TotalLength    uint64 `json:"totalLength"`
		InfoHash       []byte `json:"infoHash"`
		BytesCompleted uint64 `json:"bytesCompleted"`
		PieceLength    uint64 `json:"pieceLength"`
		Completed      bool   `json:"completed"`
	}{
		Files:          e.files,
		Name:           e.name,
		NumPieces:      e.numPieces,
		TotalLength:    e.totalLength,
		InfoHash:       e.infoHash,
		BytesCompleted: e.BytesCompleted(),
		PieceLength:    e.pieceLength,
		Completed:      e.Completed(),
	})
}

func truncate(s string, i int) string {
	runes := []rune(s)
	if len(runes) > i {
		return string(runes[:i]) + "..."
	}
	return s
}

func (e entry) String() string {
	name := truncate(e.name, 16)
	percentage := float64(e.bytesCompleted) / float64(e.totalLength) * 100.0
	completed := strings.ToLower(bytefmt.ByteSize(e.bytesCompleted))
	size := strings.ToLower(bytefmt.ByteSize(e.totalLength))
	return fmt.Sprintf("%s %s/%s %02.2f%%", name, completed, size, percentage)
}

func jsonTorrent(t *torren.Torrent) (data []byte) {
	data, _ = newTorrent(t, false).MarshalJSON()
	return
}
