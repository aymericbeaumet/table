package sfu

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
)

const CAPACITY = 256 // uint8

type roomPeerEvent struct {
	peerID uint32
	peer   *Peer
	peerEvent
}

type Room struct {
	// Immutable after init
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	eventsC chan roomPeerEvent

	// Incremented atomically when a peer is registered.
	peersCount uint32

	// Only read/write through the sequential Loop() function.
	peersByID [CAPACITY]*Peer
}

func NewRoom(name string) *Room {
	return &Room{
		ID:        uuid.New(),
		Name:      RoomNormalizeName(name),
		eventsC:   make(chan roomPeerEvent),
		peersByID: [CAPACITY]*Peer{},
	}
}

func (r *Room) RegisterPeer(peer *Peer) error {
	peersCount := atomic.AddUint32(&r.peersCount, 1)
	if peersCount > CAPACITY {
		return errors.New("room is out of capacity")
	}

	peerID := peersCount - 1
	peer.onEvent(func(event peerEvent) {
		r.eventsC <- roomPeerEvent{
			peerID:    peerID,
			peer:      peer,
			peerEvent: event,
		}
	})

	return nil
}

func (r *Room) Loop() {
	for event := range r.eventsC {
		switch {

		case event.connectionOpen != nil:
			{
				r.peersByID[event.peerID] = event.peer
				for i, max := uint32(0), r.PeersCount(); i < max; i++ {
					p := r.peersByID[i]
					p.sendPeer(event.peerID, event.peer)
					event.peer.sendPeer(i, p)
					//p.EachTrack(func(t webrtc.TrackLocal) {
					//if err := peer.AddTrack(t); err != nil {
					//fmt.Println("error:", err)
					//}
					//})
				}
			}

		case event.connectionClose != nil:
			{
				r.peersByID[event.peerID] = nil
				for i, max := uint32(0), r.PeersCount(); i < max; i++ {
					if i != event.peerID {
						p := r.peersByID[i]
						p.sendPeer(event.peerID, event.peer)
					}
				}
			}

		case event.remoteTrack != nil:
			{
				//r.EachPeer(func(p *Peer) {
				//if p != peer {
				//if err := p.AddTrack(track); err != nil {
				//fmt.Println("error:", err)
				//}
				//}
				//})
			}

		case event.reliableMessage != nil:
			{
				for i, max := uint32(0), r.PeersCount(); i < max; i++ {
					if i != event.peerID {
						p := r.peersByID[i]
						p.sendReliable(event.peerID, event.reliableMessage.data)
					}
				}
			}

		case event.unreliableMessage != nil:
			{
				for i, max := uint32(0), r.PeersCount(); i < max; i++ {
					if i != event.peerID {
						p := r.peersByID[i]
						p.sendUnreliable(event.peerID, event.unreliableMessage.data)
					}
				}
			}

		default:
			panic(fmt.Errorf("implementation error: unsupported peer event: %#v", event))

		}
	}
}

func (r *Room) PeersCount() uint32 {
	return atomic.LoadUint32(&r.peersCount)
}

func (r *Room) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":       r.ID,
		"name":     r.Name,
		"used":     r.PeersCount(),
		"capacity": CAPACITY,
	})
}

func RoomNormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
