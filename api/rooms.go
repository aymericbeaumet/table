package main

import (
	"encoding/json"
	"sync"
)

type Rooms struct {
	roomsByName map[string]*Room
	mu          sync.RWMutex
}

func NewRooms() *Rooms {
	return &Rooms{
		roomsByName: map[string]*Room{},
	}
}

func (r *Rooms) Get(name string) *Room {
	name = NormalizeRoomName(name)

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.roomsByName[name]
}

func (r *Rooms) Upsert(name string) *Room {
	name = NormalizeRoomName(name)

	if room := r.Get(name); room != nil {
		return room
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	room, ok := r.roomsByName[name]
	if !ok {
		room = NewRoom(name)
		r.roomsByName[name] = room
	}

	return room
}

func (r *Rooms) MarshalJSON() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return json.Marshal(r.roomsByName)
}
