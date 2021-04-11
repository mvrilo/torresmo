package http

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"github.com/mvrilo/torresmo/errors"
	"github.com/mvrilo/torresmo/log"
	"github.com/mvrilo/torresmo/stream"
	"github.com/mvrilo/torresmo/torrent"
)

type handler struct {
	*gin.Engine
	stream stream.Publisher
	client torrent.Client
	logger log.Logger
}

var _ http.Handler = &handler{}

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
		ctx.Error(errors.ErrBadRequest.Wrap(err))
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

// error handling middleware
func (h *handler) errorHandlingMiddleware(ctx *gin.Context) {
	ctx.Next()

	err := ctx.Errors.Last()
	if err == nil {
		return
	}

	status := http.StatusInternalServerError
	var body interface{} = http.StatusText(status)

	switch er := err.Err.(type) {
	case errors.Error:
		ctx.AbortWithStatusJSON(er.Status, er.Error())
	default:
		ctx.AbortWithStatusJSON(status, body)
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
	} else {
		log.Error("Request")
	}
}

// NewHandler returns a default handler implemented with gin
func NewHandler(cli torrent.Client, logger log.Logger, files fs.FS, publisher stream.Publisher, debug bool) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	h := &handler{
		Engine: router,
		stream: publisher,
		client: cli,
		logger: logger,
	}

	if debug {
		ginpprof.Wrap(router)
	}

	router.Use(h.loggingMiddleware)
	router.Use(h.errorHandlingMiddleware)
	router.GET("/api/stats/", h.Stats)
	router.GET("/api/torrents/", h.Torrents)
	router.POST("/api/torrents/", h.AddTorrent)
	router.GET("/api/events/", gin.WrapH(h.stream))
	router.Use(gin.WrapH(http.FileServer(http.FS(files))))

	return h
}
