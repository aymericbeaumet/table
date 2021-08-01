package sfu

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/vmihailenco/msgpack/v5"
)

type envelope struct {
	PeerID uint32             `msgpack:"peerID"`
	Data   msgpack.RawMessage `msgpack:"data"`
}

type Peer struct {
	CreatedAt               time.Time `msgpack:"created_at" json:"created_at"`
	onEventFunc             func(peerEvent)
	onNegotiationNeededFunc func()

	peerConnection        *webrtc.PeerConnection
	reliableDataChannel   *webrtc.DataChannel
	unreliableDataChannel *webrtc.DataChannel
}

func NewPeer() (*Peer, error) {
	p := &Peer{
		CreatedAt: time.Now().UTC(),
	}

	// Create a new RTCPeerConnection
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: strings.Split(os.Getenv("STUN_SERVERS_URLS"), ","),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	p.peerConnection = pc

	// Create pre-negotiated data channels

	rNegotiated, rID, rOrdered := true, uint16(0), false
	reliableDataChannel, err := p.peerConnection.CreateDataChannel("reliable", &webrtc.DataChannelInit{
		Negotiated: &rNegotiated,
		ID:         &rID,
		//
		Ordered: &rOrdered,
	})
	if err != nil {
		return nil, err
	}
	reliableDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		p.onEventFunc(peerEvent{reliableMessage: &peerEventReliableMessage{msg.Data}})
	})
	p.reliableDataChannel = reliableDataChannel

	uNegotiated, uID, uOrdered, uMaxRestransmits := true, uint16(1), false, uint16(0)
	unreliableDataChannel, err := p.peerConnection.CreateDataChannel("unreliable", &webrtc.DataChannelInit{
		Negotiated: &uNegotiated,
		ID:         &uID,
		//
		Ordered:        &uOrdered,
		MaxRetransmits: &uMaxRestransmits,
	})
	if err != nil {
		return nil, err
	}
	unreliableDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		p.onEventFunc(peerEvent{unreliableMessage: &peerEventUnreliableMessage{msg.Data}})
	})
	p.unreliableDataChannel = unreliableDataChannel

	p.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case
			webrtc.PeerConnectionStateConnected:
			{
				p.onEventFunc(peerEvent{connectionOpen: &peerEventConnectionOpen{}})
			}

		case
			webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateFailed,
			webrtc.PeerConnectionStateClosed:
			{
				p.onEventFunc(peerEvent{connectionClose: &peerEventConnectionClose{}})
			}
		}
	})

	p.peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		p.onEventFunc(peerEvent{remoteTrack: &peerEventRemoteTrack{remoteTrack}})
	})

	return p, nil
}

func (p *Peer) onEvent(f func(event peerEvent)) {
	p.onEventFunc = f
}

func (p *Peer) sendPeer(sourcePeerID uint32, peer *Peer) error {
	if p == nil {
		return nil
	}
	msg, err := msgpack.Marshal(map[string]interface{}{
		"PEER": peer,
	})
	if err != nil {
		return err
	}
	return p.sendReliable(sourcePeerID, msg)
}

func (p *Peer) sendReliable(sourcePeerID uint32, data []byte) error {
	if p == nil {
		return nil
	}
	wrapped, err := p.wrap(sourcePeerID, data)
	if err != nil {
		return err
	}
	return p.send(p.reliableDataChannel, wrapped)
}

func (p *Peer) sendUnreliable(sourcePeerID uint32, data []byte) error {
	if p == nil {
		return nil
	}
	wrapped, err := p.wrap(sourcePeerID, data)
	if err != nil {
		return err
	}
	return p.send(p.unreliableDataChannel, wrapped)
}

func (p *Peer) send(dc *webrtc.DataChannel, data []byte) error {
	if dc == nil {
		return errors.New("channel does not exist")
	}
	return dc.Send(data)
}

func (p *Peer) wrap(sourcePeer uint32, data []byte) ([]byte, error) {
	e := &envelope{
		PeerID: sourcePeer,
		Data:   data,
	}
	return msgpack.Marshal(e)
}

func (p *Peer) OnICECandidate(f func(candidate *webrtc.ICECandidate)) {
	p.peerConnection.OnICECandidate(f)
}

func (p *Peer) OnNegotiationNeeded(f func()) {
	p.onNegotiationNeededFunc = f
	p.peerConnection.OnNegotiationNeeded(f)
}

func (p *Peer) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return p.peerConnection.AddICECandidate(candidate)
}

func (p *Peer) Handshake(remoteDesc webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if remoteDesc.Type != webrtc.SDPTypeOffer {
		return nil, fmt.Errorf("unexpected session description of type %#v", remoteDesc.Type)
	}

	if err := p.peerConnection.SetRemoteDescription(remoteDesc); err != nil {
		return nil, err
	}

	answer, err := p.peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err := p.peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return p.peerConnection.LocalDescription(), nil
}
