package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mvrilo/torresmo/log"
)

type wsPublisher struct {
	rooms map[string]map[net.Conn]struct{}
	mu    *sync.Mutex
	log   log.Logger
}

func (s *wsPublisher) addConn(room string, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rooms[room]; !ok {
		s.rooms[room] = make(map[net.Conn]struct{})
		s.rooms[room][conn] = struct{}{}
		s.log.Info("ws: new connection on", room)
	}
}

func (s *wsPublisher) getRooms() (rooms []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for room := range s.rooms {
		rooms = append(rooms, room)
	}
	return
}

func (s *wsPublisher) getRoomConns(room string) (conns []net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for roomConn := range s.rooms[room] {
		conns = append(conns, roomConn)
	}
	return
}

func (s *wsPublisher) getConns() (conns []net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, roomConns := range s.rooms {
		for conn := range roomConns {
			conns = append(conns, conn)
		}
	}
	return
}

func (s *wsPublisher) closeRoomConn(room string, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, roomConns := range s.rooms {
		if _, ok := roomConns[conn]; ok {
			conn.Close()
			delete(roomConns, conn)
			s.log.Info("ws: closed connection on", room)
		}
	}
	return
}

func (s *wsPublisher) handleConn(rooms []string, conn net.Conn) {
	for _, room := range rooms {
		s.addConn(room, conn)
		// defer s.closeRoomConn(room, conn)
	}

	for {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil && err != io.EOF {
			// s.log.Error("websocket error:", errors.Wrap(err, "websocket read"))
			continue
		}

		if op == ws.OpContinuation {
			continue
		}

		fmt.Printf("ws: read: %+v %+v\n", op, msg)
	}
}

func (s *wsPublisher) Serve() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqRooms := r.URL.Query().Get("rooms")
		rooms := strings.Split(reqRooms, ",")
		if len(rooms) < 1 {
			w.WriteHeader(400)
			w.Write([]byte(`{ "error": "missing \"rooms\" parameter" }`))
			return
		}

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(`{ "error": "` + err.Error() + `" }`))
			return
		}

		go s.handleConn(rooms, conn)
	}
}

func (s *wsPublisher) Broadcast(data []byte) {
	for _, conn := range s.getConns() {
		err := wsutil.WriteServerMessage(conn, ws.OpText, data)
		if err != nil {
			continue
		}
	}
	return
}

func (s *wsPublisher) Publish(room string, data interface{}) {
	reply, err := json.Marshal(struct {
		Room    string      `json:"room"`
		Torrent interface{} `json:"torrent"`
	}{
		room, data,
	})
	if err != nil {
		return
	}

	for _, conn := range s.getRoomConns(room) {
		err := wsutil.WriteServerMessage(conn, ws.OpText, reply)
		if err != nil {
			continue
		}
	}
	return
}

// NewWebsocket returns a websocket implementation of Publisher
func NewWebsocket(log log.Logger) Publisher {
	return &wsPublisher{
		rooms: make(map[string]map[net.Conn]struct{}),
		log:   log,
		mu:    new(sync.Mutex),
	}
}
