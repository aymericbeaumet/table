package main

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	name string

	mu    sync.Mutex
	count int
	conns map[int]*websocket.Conn
}

func NewRoom(name string) *Room {
	return &Room{
		name:  name,
		conns: map[int]*websocket.Conn{},
	}
}

func (r *Room) GetConn(id int) (*websocket.Conn, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.conns[id]
	return v, ok
}

func (r *Room) EachConn(cb func(int, *websocket.Conn)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range r.conns {
		cb(k, v)
	}
}

func (r *Room) AddConn(conn *websocket.Conn) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count += 1
	r.conns[r.count] = conn
	return r.count
}

func (r *Room) DelConn(id int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.conns, id)
}

func (r *Room) MarshalJSON() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return json.Marshal(r.conns)
}
