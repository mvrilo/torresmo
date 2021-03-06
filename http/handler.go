package http

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mvrilo/torresmo/errors"
	"github.com/mvrilo/torresmo/event"
	"github.com/mvrilo/torresmo/log"
	"github.com/mvrilo/torresmo/torrent"
)

type handler struct {
	*gin.Engine
	eventHandler event.Handler
	client       torrent.Client
	logger       log.Logger
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
	latency := time.Since(start)

	log = log.With(
		"status", status,
		"method", req.Method,
		"path", path,
		"ip", ctx.ClientIP(),
		"latency", latency.String(),
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
func NewHandler(cli torrent.Client, logger log.Logger, staticFiles fs.FS, downloadedFiles fs.FS, eventHandler event.Handler, debug bool) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	h := &handler{
		Engine:       router,
		eventHandler: eventHandler,
		client:       cli,
		logger:       logger,
	}

	if debug {
		ginpprof.Wrap(router)
	}

	router.Use(cors.Default())
	router.Use(h.loggingMiddleware)
	router.Use(h.errorHandlingMiddleware)
	if downloadedFiles != nil {
		router.Any("/downloads/*any", gin.WrapH(http.StripPrefix("/downloads/", http.FileServer(http.FS(downloadedFiles)))))
	}
	router.GET("/api/stats/", h.Stats)
	router.GET("/api/torrents/", h.Torrents)
	router.POST("/api/torrents/", h.AddTorrent)
	router.GET("/api/events/", gin.WrapH(h.eventHandler.(http.Handler)))
	router.Use(gin.WrapH(http.FileServer(http.FS(staticFiles))))

	return h
}
