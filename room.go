package main

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type Room struct {
	ID    string              `json:"id"`
	Peers map[uuid.UUID]*Peer `json:"peers"`
	mu    sync.Mutex
}

func NewRoom(name string) *Room {
	return &Room{
		ID:    slug.Make(name),
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
