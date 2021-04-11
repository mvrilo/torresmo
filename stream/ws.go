package stream

import (
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mvrilo/torresmo/log"
)

var _ Publisher = &wsPublisher{}

type wsPublisher struct {
	conns map[net.Conn]struct{}
	mu    *sync.Mutex
	log   log.Logger
}

func (s *wsPublisher) closeConn(conn net.Conn) {
	s.mu.Lock()
	conn.Close()
	if _, ok := s.conns[conn]; ok {
		delete(s.conns, conn)
	}
	s.mu.Unlock()
}

func (s *wsPublisher) handleConn(conn net.Conn) {
	s.mu.Lock()
	if _, ok := s.conns[conn]; !ok {
		s.conns[conn] = struct{}{}
	}
	s.mu.Unlock()

	// defer s.closeConn(conn)
	// for {
	// 	msg, op, err := wsutil.ReadClientData(conn)
	// 	if err != nil && err != io.EOF {
	// 		// s.log.Error("websocket error:", errors.Wrap(err, "websocket read"))
	// 		continue
	// 	}
	// 	if op == ws.OpContinuation {
	// 		continue
	// 	}
	// 	s.log.Info(op, msg)
	// }
}

func (s *wsPublisher) getConns() (conns []net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.conns {
		conns = append(conns, conn)
	}
	return
}

func (s *wsPublisher) removeConn(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conn.Close()
	delete(s.conns, conn)
}

func (s *wsPublisher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(`{ "error": "` + err.Error() + `" }`))
		return
	}
	go s.handleConn(conn)
}

func (s *wsPublisher) Publish(data []byte) {
	for _, conn := range s.getConns() {
		err := wsutil.WriteServerMessage(conn, ws.OpText, data)
		if err != nil {
			continue
		}
	}
	return
}

func Websocket(logger log.Logger) Publisher {
	return &wsPublisher{
		conns: make(map[net.Conn]struct{}),
		mu:    new(sync.Mutex),
		log:   logger,
	}
}
