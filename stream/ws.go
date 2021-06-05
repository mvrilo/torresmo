package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mvrilo/torresmo/log"
)

type wsPublisher struct {
	online uint
	topics map[fmt.Stringer]map[net.Conn]struct{}
	mu     *sync.Mutex
	log    log.Logger
}

var _ Publisher = (*wsPublisher)(nil)

func (s *wsPublisher) removeConn(topic Topic, conn net.Conn) {
	s.mu.Lock()
	delete(s.topics[topic], conn)
	s.mu.Unlock()
}

func (s *wsPublisher) addConn(topic Topic, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.topics[topic]; !ok {
		s.topics[topic] = make(map[net.Conn]struct{})
	}

	s.topics[topic][conn] = struct{}{}
	s.log.Info("ws: new connection on topic: ", topic)
}

func (s *wsPublisher) getTopicConns(topic Topic) (conns []net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for topicConn := range s.topics[topic] {
		conns = append(conns, topicConn)
	}
	return
}

func (s *wsPublisher) handleConn(topics []Topic, conn net.Conn) {
	for _, topic := range topics {
		s.addConn(topic, conn)
	}
	s.online++
	s.Publish(TopicOnline, s.online)

	for {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil && err == io.EOF {
			continue
		}

		if err != nil {
			for _, topic := range topics {
				s.removeConn(topic, conn)
			}
			s.online--
			s.Publish(TopicOnline, s.online)
			conn.Close()
			break
		}

		if op == ws.OpContinuation {
			continue
		}

		fmt.Printf("ws: read: %+v %+v\n", op, msg)
	}
}

// Serve retrns a http.HandlerFunc
func (s *wsPublisher) Serve() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(`{ "error": "` + err.Error() + `" }`))
			return
		}

		go s.handleConn(AllTopics, conn)
	}
}

// Publish sends data to a topic
func (s *wsPublisher) Publish(topic Topic, data interface{}) {
	reply, err := json.Marshal(struct {
		Topic string      `json:"topic"`
		Data  interface{} `json:"data"`
	}{
		topic.String(), data,
	})
	if err != nil {
		return
	}

	for _, conn := range s.getTopicConns(topic) {
		err := wsutil.WriteServerMessage(conn, ws.OpText, reply)
		if err != nil {
			continue
		}
	}
}

// NewWebsocket returns a websocket implementation of Publisher
func NewWebsocket(log log.Logger) Publisher {
	return &wsPublisher{
		topics: make(map[fmt.Stringer]map[net.Conn]struct{}),
		log:    log,
		mu:     new(sync.Mutex),
	}
}
