package event

import (
	"net"
	"net/http"
)

type discardHandler struct{}

var _ Handler = discardHandler{}

func (discardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
func (discardHandler) Publish(Topic, interface{})                       {}
func (discardHandler) Subscribe(Topic, net.Conn)                        {}
func (discardHandler) Unsubscribe(Topic, net.Conn)                      {}

// Discard handler
func Discard() discardHandler {
	return discardHandler{}
}
