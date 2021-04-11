package stream

import "net/http"

type Publisher interface {
	http.Handler
	Publish(data []byte)
}
