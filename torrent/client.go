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
	torren "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"golang.org/x/time/rate"

	"github.com/mvrilo/torresmo/event"
	"github.com/mvrilo/torresmo/log"
)

const MimeBitTorrent = "application/x-bittorrent"

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
	WithEventHandler(pub event.Handler) Client
	WithDownloadLimit(limit int) Client
	WithUploadLimit(limit int) Client
	WithTorrentFiles(dir string) Client
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
	OutputDir() string
}

type client struct {
	*torren.Client
	conf *torren.ClientConfig

	Logger       log.Logger
	eventHandler event.Handler

	mu           sync.Mutex
	watchDir     string
	dumpDir      string
	biggestFirst bool
}

func NewClient(logger log.Logger) (Client, error) {
	cli := &client{
		Logger: logger,
		conf:   torren.NewDefaultClientConfig(),
		mu:     sync.Mutex{},
	}

	cli.conf.Logger = log2.Discard
	return cli, nil
}

func (c *client) OutputDir() string {
	return c.conf.DataDir
}

func (c *client) Start() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Logger.Info("Starting Torresmo torrent client")
	c.Client, err = torren.NewClient(c.conf)
	return
}

func (c *client) WithEventHandler(pub event.Handler) Client {
	c.eventHandler = pub
	return c
}

func (c *client) WithSeed(seed bool) Client {
	c.conf.NoUpload = !seed
	return c
}

func (c *client) WithOutput(output string) Client {
	c.conf.DataDir = output
	c.conf.DefaultStorage = storage.NewMMapWithCompletion(output, storage.NewMapPieceCompletion())
	return c
}

func (c *client) WithTorrentFiles(dir string) Client {
	c.dumpDir = dir
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
	go func() {
		<-t.GotInfo()
		t.DownloadAll()
		c.writeTorrentFile(t)

		nt := newTorrent(t)
		ch <- nt

		evthandler := c.eventHandler
		evthandler.Publish(event.TopicStarted, nt)

		if c.biggestFirst {
			BiggestFileFromTorrent(nt).Now()
		}

		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			<-ticker.C
			nt = newTorrent(t)
			evthandler.Publish(event.TopicDownloading, nt)
			if nt.Completed() {
				evthandler.Publish(event.TopicCompleted, nt)
			}
		}
	}()
	return ch
}

func (c *client) getTorrentFilesFilename() (filenames []string) {
	if c.watchDir != "" {
		files, err := filepath.Glob(c.watchDir + "/*.torrent")
		if err != nil {
			c.Logger.Error("error getting files")
			return
		}
		filenames = append(filenames, files...)
	}
	return
}

func (c *client) ReadTorrentFiles() error {
	files := c.getTorrentFilesFilename()
	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("error opening file: %s", name))
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

func (c *client) writeTorrentFile(t *torren.Torrent) error {
	err := os.MkdirAll(c.dumpDir, 0750)
	if err != nil {
		return err
	}

	path := filepath.Join(c.dumpDir, t.Name()+".torrent")
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
		if bytes.Equal([]byte(t.InfoHash()), hash) {
			return t
		}
	}

	if t, ok := c.Client.Torrent(metainfo.HashBytes(hash)); ok {
		return newTorrent(t)
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
	spec, err := torren.TorrentSpecFromMagnetUri(uri)
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

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != MimeBitTorrent {
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
		torrents = append(torrents, newTorrent(t))
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
