package main

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID           `json:"id"`
	Name      string              `json:"name"`
	PeersByID map[uuid.UUID]*Peer `json:"peersByID"`
	mu        sync.RWMutex
}

func NewRoom(name string) *Room {
	return &Room{
		ID:        uuid.New(),
		Name:      NormalizeRoomName(name),
		PeersByID: map[uuid.UUID]*Peer{},
	}
}

func (r *Room) GetPeer(id uuid.UUID) *Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.PeersByID[id]
}

func (r *Room) EachPeer(cb func(*Peer)) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, peer := range r.PeersByID {
		cb(peer)
	}
}

func (r *Room) AddPeer(peer *Peer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.PeersByID[peer.ID] = peer
	return nil
}

func (r *Room) DelPeer(peer *Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.PeersByID, peer.ID)
}

func NormalizeRoomName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
