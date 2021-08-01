package sfu

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	webrtc "github.com/pion/webrtc/v3"
	"github.com/vmihailenco/msgpack/v5"
)

const CAPACITY = 256 // uint8

type roomPeerEvent struct {
	peerID uint32
	peer   *Peer
	peerEvent
}

type roomPeerStatus struct {
	PEER roomPeerStatusPeer `msgpack:"PEER"`
}

type roomPeerStatusPeer struct {
	Status string `msgpack:"status"`
}

func newRoomPeerStatus(status string) *roomPeerStatus {
	return &roomPeerStatus{
		PEER: roomPeerStatusPeer{
			Status: status,
		},
	}
}

var statusOnline, statusOffline = newRoomPeerStatus("online"), newRoomPeerStatus("offline")

type Room struct {
	// Immutable after init
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	eventsC chan roomPeerEvent

	// Only read/write through the sequential Loop() function.
	peersByID map[uint32]*Peer

	// Incremented atomically when a peer is registered.
	nextPeerID uint32
}

func NewRoom(name string) *Room {
	return &Room{
		ID:        uuid.New(),
		Name:      RoomNormalizeName(name),
		eventsC:   make(chan roomPeerEvent, 1024),
		peersByID: make(map[uint32]*Peer, 16),
	}
}

func (r *Room) RegisterPeer(peer *Peer) {
	peerID := atomic.AddUint32(&r.nextPeerID, 1)
	peer.onEvent(func(event peerEvent) {
		r.eventsC <- roomPeerEvent{
			peerID:    peerID,
			peer:      peer,
			peerEvent: event,
		}
	})
}

func (r *Room) Loop() {
	for event := range r.eventsC {
		switch {

		case event.connectionOpen != nil:
			{
				r.peersByID[event.peerID] = event.peer
				for id, peer := range r.peersByID {
					if id != event.peerID {
						d1, _ := msgpack.Marshal(statusOnline)
						peer.sendReliable(event.peerID, d1)

						d2, _ := msgpack.Marshal(statusOnline)
						event.peer.sendReliable(id, d2)

						//peer.EachTrack(func(t webrtc.TrackLocal) {
						//if err := peer.AddTrack(t); err != nil {
						//fmt.Println("error:", err)
						//}
						//})
					}
				}
			}

		case event.connectionClose != nil:
			{
				delete(r.peersByID, event.peerID)
				for _, peer := range r.peersByID {
					d, _ := msgpack.Marshal(statusOffline)
					peer.sendReliable(event.peerID, d)
				}
			}

		case event.remoteTrack != nil:
			{
				localTrack, err := consumeRemoteTrack(event.peerID, event.remoteTrack.TrackRemote)
				if err != nil {
					fmt.Println("error:", err)
				}

				for id, peer := range r.peersByID {
					if id != event.peerID {
						if err := peer.addTrack(localTrack); err != nil {
							fmt.Println("error:", err)
						}
					}
				}
			}

		case event.reliableMessage != nil:
			{
				for id, peer := range r.peersByID {
					if id != event.peerID {
						peer.sendReliable(event.peerID, event.reliableMessage.data)
					}
				}
			}

		case event.unreliableMessage != nil:
			{
				for id, peer := range r.peersByID {
					if id != event.peerID {
						peer.sendUnreliable(event.peerID, event.unreliableMessage.data)
					}
				}
			}

		default:
			panic(fmt.Errorf("implementation error: unsupported peer event: %#v", event))

		}
	}
}

func RoomNormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func consumeRemoteTrack(peerID uint32, remoteTrack *webrtc.TrackRemote) (webrtc.TrackLocal, error) {
	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		remoteTrack.Codec().RTPCodecCapability,
		remoteTrack.Kind().String(),
		strconv.FormatUint(uint64(peerID), 10),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		rtpBuf := make([]byte, 1400)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			if readErr != nil {
				panic(readErr)
			}

			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err = localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				panic(err)
			}
		}
	}()

	return localTrack, nil
}
