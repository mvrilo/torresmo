package event

import (
	"net"
)

type Handler interface {
	Publish(topic Topic, data interface{})
	Subscribe(topic Topic, conn net.Conn)
	Unsubscribe(topic Topic, conn net.Conn)
}
