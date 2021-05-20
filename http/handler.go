package http

import (
	"context"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"github.com/mvrilo/torresmo/cast"
	"github.com/mvrilo/torresmo/errors"
	"github.com/mvrilo/torresmo/log"
	"github.com/mvrilo/torresmo/stream"
	"github.com/mvrilo/torresmo/torrent"
)

type handler struct {
	*gin.Engine
	cast   *cast.Cast
	stream stream.Publisher
	client torrent.Client
	logger log.Logger
}

var _ http.Handler = (*handler)(nil)

type AddTorrentRequest struct {
	URI string `json:"uri" binding:"required,uri"`
}

func (h *handler) AddTorrent(ctx *gin.Context) {
	var params AddTorrentRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		ctx.Error(errors.ErrBadRequest.Wrap(err))
		return
	}

	torrent, err := h.client.AddURI(params.URI)
	if err != nil {
		ctx.Error(errors.ErrInternal.Wrap(err))
		return
	}

	ctx.JSON(http.StatusCreated, <-torrent)
}

func (h *handler) Torrents(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.client.Torrents())
}

func (h *handler) Stats(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.client.Stats())
}

func (h *handler) ListDevices(ctx *gin.Context) {
	t, err := strconv.Atoi(ctx.DefaultQuery("t", "1"))
	if err != nil {
		t = 1
	}
	ct, cancel := context.WithTimeout(ctx, time.Duration(t)*time.Second)
	defer cancel()

	devices, err := h.cast.Devices(ct)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, devices)
}

func (h *handler) ConnectDevice(ctx *gin.Context) {
	ct, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := h.cast.Connect(ct, ctx.Param("uuid"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (h *handler) play(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	contentType, err := fileContentType(file)
	if err != nil {
		return err
	}

	return h.cast.Application.Load(filepath, contentType, true, true, true)
}

func (h *handler) PlayHash(ctx *gin.Context) {
	hash := ctx.Param("hash")
	for _, t := range h.client.Torrents() {
		if hash != t.InfoHash() {
			continue
		}

		f := torrent.BiggestFileFromTorrent(t)
		outputdir := filepath.Join(h.client.OutputDir(), f.Path())
		err := h.play(outputdir)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.Status(http.StatusCreated)
		break
	}
}

func (h *handler) StopCast(ctx *gin.Context) {
	h.cast.Application.Stop()
	ctx.Status(http.StatusNoContent)
}

func (h *handler) LoadFile(ctx *gin.Context) {
	filepath := ctx.Query("filepath")
	if filepath == "" {
		ctx.Error(errors.New("filepath missing"))
		return
	}

	err := h.play(filepath)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *handler) CastStatus(ctx *gin.Context) {
	_, media, volume := h.cast.Application.Status()
	ctx.JSON(http.StatusOK, struct {
		Media  interface{} `json:"media"`
		Volume interface{} `json:"volume"`
	}{
		Media:  media,
		Volume: volume,
	})
}

// error handling middleware
func (h *handler) errorHandlingMiddleware(ctx *gin.Context) {
	ctx.Next()

	err := ctx.Errors.Last()
	if err == nil {
		return
	}

	switch er := err.Err.(type) {
	case *errors.Error:
		ctx.AbortWithStatusJSON(er.Code, er.Error())
	default:
	}
}

// custom logging log middleware
func (h *handler) loggingMiddleware(ctx *gin.Context) {
	start := time.Now()
	log := h.logger
	req := ctx.Request
	path := req.URL.Path
	if raw := req.URL.RawQuery; raw != "" {
		path = path + "?" + raw
	}

	ctx.Next()

	w := ctx.Writer
	status := w.Status()
	latency := time.Now().Sub(start)

	log = log.With(
		"status", status,
		"method", req.Method,
		"path", path,
		"ip", ctx.ClientIP(),
		"latency", latency,
		"latencyhuman", latency.String(),
		"user-agent", req.UserAgent(),
	)

	if status == http.StatusNotFound || status == http.StatusUnauthorized || status < http.StatusBadRequest {
		log.Info("Request")
		return
	}

	err := ctx.Errors.Last()
	if err != nil {
		log.Error(err)
	}
}

// NewHandler returns a default handler implemented with gin
func NewHandler(cli torrent.Client, logger log.Logger, files fs.FS, publisher stream.Publisher, cast *cast.Cast, debug bool) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	h := &handler{
		Engine: router,
		stream: publisher,
		client: cli,
		logger: logger,
		cast:   cast,
	}

	if debug {
		ginpprof.Wrap(router)
	}

	router.Use(h.loggingMiddleware)
	router.Use(h.errorHandlingMiddleware)

	router.GET("/api/stats/", h.Stats)
	router.GET("/api/torrents/", h.Torrents)
	router.POST("/api/torrents/", h.AddTorrent)

	router.GET("/api/cast/status/", h.CastStatus)
	router.GET("/api/cast/devices/", h.ListDevices)
	router.GET("/api/cast/connect/:uuid", h.ConnectDevice)
	router.POST("/api/cast/load/", h.LoadFile)
	router.POST("/api/cast/play/:hash", h.PlayHash)
	router.POST("/api/cast/stop/", h.StopCast)

	router.GET("/api/events/", gin.WrapH(h.stream))

	router.Use(gin.WrapH(http.FileServer(http.FS(files))))

	return h
}

func fileContentType(out *os.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}
