package main

import (
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

var roomID uint32

type Room struct {
	ID        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	PeersByID map[uint32]*Peer `json:"peersByID"`
	mu        sync.RWMutex
	inc       uint32
}

func NewRoom(name string) *Room {
	return &Room{
		ID:        uuid.New(),
		Name:      NormalizeRoomName(name),
		PeersByID: map[uint32]*Peer{},
	}
}

func (r *Room) GetPeer(id uint32) *Peer {
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

func (r *Room) AddPeer(peer *Peer) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.PeersByID[peer.ID]
	r.PeersByID[peer.ID] = peer
	return !ok
}

func (r *Room) DelPeer(peer *Peer) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.PeersByID[peer.ID]
	delete(r.PeersByID, peer.ID)
	return ok
}

func (r *Room) Inc() uint32 {
	return atomic.AddUint32(&r.inc, 1)
}

func NormalizeRoomName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

type Room_ Room

func (r *Room) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return json.Marshal((*Room_)(r))
}
