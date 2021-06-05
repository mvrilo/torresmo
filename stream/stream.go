package stream

import "net/http"

type Publisher interface {
	Serve() http.HandlerFunc
	Publish(topic Topic, data interface{})
}
