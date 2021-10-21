package torrent

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	torren "github.com/anacrolix/torrent"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Torrent interface {
	Files() []File
	InfoHash() string
	Name() string
	TotalLength() uint64
	URI() string
	BytesCompleted() uint64
	Completed() bool
	MarshalJSON() ([]byte, error)
	String() string
}

type torrent struct {
	t              *torren.Torrent
	uri            string
	name           string
	numPieces      int
	files          []File
	nodes          []string
	infoHash       string
	totalLength    uint64
	bytesCompleted uint64
	pieceLength    uint64
}

func BiggestFileFromTorrent(t Torrent) (biggest File) {
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

func newTorrent(t *torren.Torrent) Torrent {
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
			files = append(files, newFile(f))
		}
	}

	metainfo := t.Metainfo()
	for _, node := range metainfo.Nodes {
		nodes = append(nodes, string(node))
	}

	return torrent{
		t:              t,
		files:          files,
		nodes:          nodes,
		name:           name,
		numPieces:      numPieces,
		totalLength:    totalLength,
		infoHash:       metainfo.HashInfoBytes().HexString(),
		bytesCompleted: uint64(t.BytesCompleted()),
		pieceLength:    pieceLength,
	}
}

func (e torrent) Files() []File          { return e.files }
func (e torrent) InfoHash() string       { return e.infoHash }
func (e torrent) Name() string           { return e.name }
func (e torrent) URI() string            { return e.uri }
func (e torrent) TotalLength() uint64    { return e.totalLength }
func (e torrent) Downloaded() int        { return e.t.NumPieces() }
func (e torrent) BytesCompleted() uint64 { return uint64(e.t.BytesCompleted()) }
func (e torrent) Completed() bool        { return e.BytesCompleted() == e.TotalLength() }

func (e torrent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		// Nodes []
		Files          []File `json:"files"`
		Name           string `json:"name"`
		NumPieces      int    `json:"numPieces"`
		TotalLength    uint64 `json:"totalLength"`
		InfoHash       string `json:"infoHash"`
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

func (e torrent) String() string {
	name := truncate(e.name, 16)
	percentage := float64(e.bytesCompleted) / float64(e.totalLength) * 100.0
	completed := strings.ToLower(bytefmt.ByteSize(e.bytesCompleted))
	size := strings.ToLower(bytefmt.ByteSize(e.totalLength))
	return fmt.Sprintf("%s %s/%s %02.2f%%", name, completed, size, percentage)
}
