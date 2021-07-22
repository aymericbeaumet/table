package main

import (
	"sync"

	"github.com/google/uuid"
	webrtc "github.com/pion/webrtc/v3"
)

var peerConnectionConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

type Peer struct {
	ID uuid.UUID

	peerConnection *webrtc.PeerConnection
}

func NewPeer(offer webrtc.SessionDescription) *Peer {
	id := uuid.New()

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		panic(err)
	}

	// Set remote description
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	return &Peer{
		ID:             id,
		peerConnection: peerConnection,
	}
}

func (p *Peer) Init() {
	// Listen for ICE candidates
	var wg sync.WaitGroup
	wg.Add(1)
	p.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			p.peerConnection.AddICECandidate(candidate.ToJSON())
		} else {
			wg.Done()
		}
	})
	wg.Wait()
}

func (p *Peer) OnDataChannel(f func(dc *webrtc.DataChannel)) {
	p.peerConnection.OnDataChannel(f)
}

func (p *Peer) OnClose(f func()) {
	p.peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateClosed || s == webrtc.PeerConnectionStateDisconnected || s == webrtc.PeerConnectionStateFailed {
			f()
		}
	})
}

func (p *Peer) Answer() *webrtc.SessionDescription {
	return p.peerConnection.LocalDescription()
}
