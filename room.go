package main

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	ID    uuid.UUID           `json:"id"`
	Name  string              `json:"name"`
	Peers map[uuid.UUID]*Peer `json:"peers"`
	mu    sync.Mutex
}

func NewRoom(name string) *Room {
	return &Room{
		ID:    uuid.New(),
		Name:  NormalizeRoomName(name),
		Peers: map[uuid.UUID]*Peer{},
	}
}

func (r *Room) GetPeer(id uuid.UUID) *Peer {
	return r.Peers[id]
}

func (r *Room) EachPeer(cb func(*Peer)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, peer := range r.Peers {
		cb(peer)
	}
}

func (r *Room) AddPeer(peer *Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Peers[peer.ID] = peer
}

func (r *Room) DelPeer(peer *Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Peers, peer.ID)
}

func NormalizeRoomName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
