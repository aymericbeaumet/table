package sfu

import (
	"encoding/json"
	"sync"
)

type SFU struct {
	// TODO: if the write lock causes congestion, partition the RoomsByName map
	// on some criteria (prefix, hash, etc).
	RoomsByName map[string]*Room `json:"roomsByName"`
	mu          sync.RWMutex
}

func New() *SFU {
	return &SFU{
		RoomsByName: map[string]*Room{},
	}
}

func (s *SFU) GetRoom(name string) *Room {
	name = RoomNormalizeName(name)

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.RoomsByName[name]
}

func (s *SFU) UpsertRoom(name string) (*Room, error) {
	name = RoomNormalizeName(name)

	if room := s.GetRoom(name); room != nil {
		return room, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.RoomsByName[name]
	if !ok {
		room = NewRoom(name)
		go room.Loop()
		s.RoomsByName[name] = room
	}
	return room, nil
}

type SFU_ SFU

func (s *SFU) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal((*SFU_)(s))
}
