package stream

import "net/http"

type discard struct{}

var _ Publisher = discard{}

func (discard) Serve() http.HandlerFunc {
	return nil
}
func (discard) Publish(Topic, interface{}) {}

func Discard() Publisher {
	return discard{}
}
