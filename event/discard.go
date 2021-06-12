package event

import (
	"net"
	"net/http"
)

type DiscardHandler struct{}

var _ Handler = DiscardHandler{}

func (DiscardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
func (DiscardHandler) Publish(Topic, interface{})                       {}
func (DiscardHandler) Subscribe(Topic, net.Conn)                        {}
func (DiscardHandler) Unsubscribe(Topic, net.Conn)                      {}
