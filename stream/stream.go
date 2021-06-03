package stream

import "net/http"

type Publisher interface {
	Serve() http.HandlerFunc
	Publish(room string, data interface{})
}
