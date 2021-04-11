package stream

import "net/http"

type discard struct{}

var _ Publisher = discard{}

func (discard) Publish([]byte) {}

func (discard) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func Discard() Publisher {
	return discard{}
}
