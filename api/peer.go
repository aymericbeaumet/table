package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/vmihailenco/msgpack/v5"
)

var peerID uint32

type Wrap struct {
	Peer *Peer              `msgpack:"peer"`
	Data msgpack.RawMessage `msgpack:"data"`
}

type Peer struct {
	ID uint32 `msgpack:"id"`

	peerConnection        *webrtc.PeerConnection
	reliableDataChannel   *webrtc.DataChannel
	unreliableDataChannel *webrtc.DataChannel
	onReliable            func([]byte)
	onUnreliable          func([]byte)
}

func NewPeer() *Peer {
	return &Peer{
		ID: atomic.AddUint32(&peerID, 1),
	}
}

func (p *Peer) Init(offer webrtc.SessionDescription) error {
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: strings.Split(os.Getenv("STUN_SERVERS_URLS"), ","),
			},
		},
	})
	if err != nil {
		return err
	}
	p.peerConnection = peerConnection

	// Set remote description
	if err := p.peerConnection.SetRemoteDescription(offer); err != nil {
		return err
	}

	// Create answer
	answer, err := p.peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err = p.peerConnection.SetLocalDescription(answer); err != nil {
		return err
	}

	// Listen for channels
	p.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		switch l := dc.Label(); l {
		case "reliable":
			p.reliableDataChannel = dc
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				if p.onReliable != nil {
					p.onReliable(msg.Data)
				}
			})
		case "unreliable":
			p.unreliableDataChannel = dc
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				if p.onUnreliable != nil {
					p.onUnreliable(msg.Data)
				}
			})
		default:
			panic(fmt.Errorf("unsupported data channel %#v", l))
		}
	})

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

	return nil
}

func (p *Peer) SendReliable(data []byte) error {
	wrapped, err := p.wrap(data)
	if err != nil {
		return err
	}
	return p.send(p.reliableDataChannel, wrapped)
}

func (p *Peer) SendUnreliable(data []byte) error {
	wrapped, err := p.wrap(data)
	if err != nil {
		return err
	}
	return p.send(p.unreliableDataChannel, wrapped)
}

func (p *Peer) OnReliable(f func(msg []byte)) {
	p.onReliable = f
}

func (p *Peer) OnUnreliable(f func(msg []byte)) {
	p.onUnreliable = f
}

func (p *Peer) OnClose(f func()) {
	p.peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateClosed || s == webrtc.PeerConnectionStateDisconnected || s == webrtc.PeerConnectionStateFailed {
			f()
		}
	})
}

func (p *Peer) LocalDescription() *webrtc.SessionDescription {
	return p.peerConnection.LocalDescription()
}

func (p *Peer) send(dc *webrtc.DataChannel, data []byte) error {
	if dc == nil {
		return errors.New("channel does not exist")
	}
	return dc.Send(data)
}

func (p *Peer) wrap(data []byte) ([]byte, error) {
	w := &Wrap{
		Peer: p,
		Data: data,
	}
	return msgpack.Marshal(w)
}
