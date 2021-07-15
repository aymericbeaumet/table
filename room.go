package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID           `json:"id"`
	Name      string              `json:"name"`
	PeersByID map[uuid.UUID]*Peer `json:"peersByID"`
	mu        sync.RWMutex

	capacity int
}

func NewRoom(name string, capacity int) *Room {
	return &Room{
		ID:        uuid.New(),
		Name:      NormalizeRoomName(name),
		PeersByID: map[uuid.UUID]*Peer{},

		capacity: capacity,
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

	if len(r.PeersByID) >= r.capacity {
		return fmt.Errorf("reached capacity of %d", r.capacity)
	}

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
