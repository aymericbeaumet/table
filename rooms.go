package main

import (
	"encoding/json"
	"sync"
)

type Rooms struct {
	mu          sync.Mutex
	roomsByName map[string]*Room
}

func NewRooms() *Rooms {
	return &Rooms{
		roomsByName: map[string]*Room{},
	}
}

func (r *Rooms) Upsert(name string) *Room {
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
