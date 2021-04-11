package torrent

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	log2 "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	torren "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"golang.org/x/time/rate"

	"github.com/mvrilo/torresmo/log"
	"github.com/mvrilo/torresmo/stream"
)

// const UserAgent = "torresmo"

var ErrTorrentAlreadyAdded = errors.New("torrent already added")

type Stats struct {
	BytesRead      uint64 `json:"bytesRead"`
	BytesWritten   uint64 `json:"bytesWritten"`
	TotalCount     uint32 `json:"totalCount"`
	CompletedCount uint32 `json:"completedCount"`
}

type ClientBuilder interface {
	WithSeed(seed bool) Client
	WithOutput(output string) Client
	WithPublisher(pub stream.Publisher) Client
	WithDownloadLimit(limit int) Client
	WithUploadLimit(limit int) Client
	WithWatchDir(dir string) Client
	WithBiggestFirst(b bool) Client
}

type Client interface {
	ClientBuilder

	Start() error
	Stop()

	AddURI(uri string) (chan Torrent, error)
	Torrents() []Torrent
	Stats() Stats
	ReadTorrentFiles() error
}

type client struct {
	*torren.Client
	conf *torren.ClientConfig

	Logger log.Logger
	stream stream.Publisher

	mu           sync.Mutex
	watchDir     string
	biggestFirst bool
}

func NewClient(logger log.Logger) (Client, error) {
	cli := &client{
		Logger: logger,
		conf:   torren.NewDefaultClientConfig(),
		mu:     sync.Mutex{},
	}

	cli.conf.Logger = log2.Discard
	// cli.conf.DisableUTP = true
	return cli, nil
}

func (c *client) Start() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Logger.Info("Starting Torresmo torrent client")
	c.Client, err = torren.NewClient(c.conf)
	return
}

func (c *client) WithPublisher(pub stream.Publisher) Client {
	c.stream = pub
	return c
}

func (c *client) WithSeed(seed bool) Client {
	c.conf.NoUpload = !seed
	return c
}

func (c *client) WithOutput(output string) Client {
	c.conf.DataDir = output
	c.conf.DefaultStorage = storage.NewMMap(output)
	return c
}

func (c *client) WithWatchDir(dir string) Client {
	c.watchDir = dir
	return c
}

func (c *client) WithDownloadLimit(limit int) Client {
	c.conf.DownloadRateLimiter = rate.NewLimiter(rate.Limit(limit*1024), limit*1024)
	return c
}

func (c *client) WithUploadLimit(limit int) Client {
	c.conf.UploadRateLimiter = rate.NewLimiter(rate.Limit(limit*1024), limit*1024)
	return c
}

func (c *client) WithBiggestFirst(b bool) Client {
	c.biggestFirst = b
	return c
}

func (c *client) Stop() {
	if c.Client != nil {
		c.Client.Close()
	}
}

func (c *client) download(t *torren.Torrent) chan Torrent {
	ch := make(chan Torrent)
	go func(t *torren.Torrent, ch chan Torrent) {
		<-t.GotInfo()
		c.writeTorrentFile(t)
		nt := newTorrent(t, c.biggestFirst)
		ch <- nt
		t.DownloadAll()

		if c.biggestFirst {
			biggestFile(nt).Now()
		}

		for {
			c.stream.Publish(jsonTorrent(t))
			<-time.After(1 * time.Second)
		}
	}(t, ch)
	return ch
}

func (c *client) getTorrentFilesFilename() []string {
	files, err := filepath.Glob(c.watchDir + "/*.torrent")
	if err != nil {
		return nil
	}
	return files
}

func (c *client) ReadTorrentFiles() error {
	files := c.getTorrentFilesFilename()
	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			continue
		}

		t, err := c.addReaderNoCheck(f)
		if err != nil {
			continue
		}

		go func(t chan Torrent) { <-t }(t)
	}
	return nil
}

func (c *client) writeTorrentFile(t *torrent.Torrent) error {
	err := os.MkdirAll(c.conf.DataDir, 0750)
	if err != nil {
		return err
	}

	path := filepath.Join(c.conf.DataDir, t.InfoHash().HexString()+".torrent")
	_, err = os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		return nil
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()
	return t.Metainfo().Write(f)
}

func (c *client) getTorrent(hash []byte) Torrent {
	for _, t := range c.Torrents() {
		if bytes.Equal(t.InfoHash(), hash) {
			return t
		}
	}

	if t, ok := c.Client.Torrent(metainfo.HashBytes(hash)); ok {
		return newTorrent(t, false)
	}

	return nil
}

func (c *client) addReaderNoCheck(r io.Reader) (chan Torrent, error) {
	metaInfo, err := metainfo.Load(r)
	if err != nil {
		return nil, err
	}

	t, err := c.Client.AddTorrent(metaInfo)
	if err != nil {
		return nil, err
	}

	return c.download(t), nil
}

func (c *client) addReader(r io.Reader) (chan Torrent, error) {
	metaInfo, err := metainfo.Load(r)
	if err != nil {
		return nil, err
	}

	if t := c.getTorrent(metaInfo.HashInfoBytes().Bytes()); t != nil {
		return nil, ErrTorrentAlreadyAdded
	}

	t, err := c.Client.AddTorrent(metaInfo)
	if err != nil {
		return nil, err
	}

	return c.download(t), nil
}

func (c *client) addMagnet(uri string) (chan Torrent, error) {
	spec, err := torrent.TorrentSpecFromMagnetUri(uri)
	if err != nil {
		return nil, err
	}

	if t := c.getTorrent(spec.InfoHash.Bytes()); t != nil {
		return nil, ErrTorrentAlreadyAdded
	}

	t, err := c.Client.AddMagnet(uri)
	if err != nil {
		return nil, err
	}
	return c.download(t), nil
}

func (c *client) httpGet(uri string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	// req.Header.Set("User-Agent", UserAgent)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "application/x-bittorrent" {
		return nil, fmt.Errorf("invalid content-type: %s", contentType)
	}

	return res.Body, nil
}

func (c *client) AddURI(uri string) (chan Torrent, error) {
	uuri, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	scheme := uuri.Scheme
	switch scheme {
	case "magnet":
		return c.addMagnet(uri)
	case "http", "https":
		body, err := c.httpGet(uri)
		if err != nil {
			return nil, err
		}

		defer body.Close()
		return c.addReader(body)
	default:
		return nil, fmt.Errorf("invalid uri scheme: %s", scheme)
	}
}

func (c *client) Torrents() (torrents []Torrent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Client == nil {
		return
	}
	for _, t := range c.Client.Torrents() {
		torrents = append(torrents, newTorrent(t, false))
	}
	return
}

func (c *client) completedTorrents() (torrents []Torrent) {
	for _, t := range c.Torrents() {
		if t.Completed() {
			torrents = append(torrents, t)
		}
	}
	return
}

func (c *client) Stats() (stats Stats) {
	st := c.Client.ConnStats()
	torrents := c.Torrents()
	completed := c.completedTorrents()
	return Stats{
		TotalCount:     uint32(len(torrents)),
		CompletedCount: uint32(len(completed)),
		BytesWritten:   uint64(st.BytesWritten.Int64()),
		BytesRead:      uint64(st.BytesRead.Int64()),
	}
}
