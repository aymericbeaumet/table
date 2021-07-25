package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	webrtc "github.com/pion/webrtc/v3"
	"github.com/vmihailenco/msgpack/v5"
)

type Wrap struct {
	PeerID uint32             `msgpack:"peerID"`
	Data   msgpack.RawMessage `msgpack:"data"`
}

type Peer struct {
	ID     uint32 `msgpack:"id"`
	Status string `msgpack:"status"`

	peerConnection        *webrtc.PeerConnection
	reliableDataChannel   *webrtc.DataChannel
	unreliableDataChannel *webrtc.DataChannel

	onOpen              func()
	onClose             func()
	onMessageReliable   func([]byte)
	onMessageUnreliable func([]byte)
	onTrack             func(*webrtc.TrackLocalStaticRTP)
}

func NewPeer(id uint32) *Peer {
	return &Peer{
		ID:     id,
		Status: "offline",

		onOpen:              func() {},
		onClose:             func() {},
		onMessageReliable:   func([]byte) {},
		onMessageUnreliable: func([]byte) {},
		onTrack:             func(*webrtc.TrackLocalStaticRTP) {},
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

	// Allow us to receive 1 video track
	if _, err = p.peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}

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

	// Listen for connection states updates
	p.peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		switch s {
		case webrtc.PeerConnectionStateConnected:
			p.Status = "online"
			p.onOpen()
		case webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateDisconnected, webrtc.PeerConnectionStateFailed:
			p.Status = "offline"
			p.onClose()
		}
	})

	// Listen for channels
	p.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		switch l := dc.Label(); l {
		case "reliable":
			p.reliableDataChannel = dc
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				p.onMessageReliable(msg.Data)
			})
		case "unreliable":
			p.unreliableDataChannel = dc
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				p.onMessageUnreliable(msg.Data)
			})
		default:
			panic(fmt.Errorf("unsupported data channel %#v", l))
		}
	})

	// Listen for tracks
	p.peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		localTrack, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", "pion")
		if err != nil {
			panic(err)
		}

		p.onTrack(localTrack)

		rtpBuf := make([]byte, 1400)
		for {
			i, _, err := remoteTrack.Read(rtpBuf)
			if err != nil {
				panic(err)
			}
			if _, err = localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				panic(err)
			}
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

func (p *Peer) SendReliable(sourcePeer *Peer, data []byte) error {
	wrapped, err := p.wrap(sourcePeer, data)
	if err != nil {
		return err
	}
	return p.send(p.reliableDataChannel, wrapped)
}

func (p *Peer) SendUnreliable(sourcePeer *Peer, data []byte) error {
	wrapped, err := p.wrap(sourcePeer, data)
	if err != nil {
		return err
	}
	return p.send(p.unreliableDataChannel, wrapped)
}

func (p *Peer) SendPeer(peer *Peer) error {
	msg, err := msgpack.Marshal(map[string]interface{}{
		"PEER": peer,
	})
	if err != nil {
		return err
	}
	return p.SendReliable(peer, msg)
}

func (p *Peer) AddTrack(track *webrtc.TrackLocalStaticRTP) error {
	_, err := p.peerConnection.AddTrack(track)
	return err
}

func (p *Peer) OnMessageReliable(f func(msg []byte)) {
	p.onMessageReliable = f
}

func (p *Peer) OnMessageUnreliable(f func(msg []byte)) {
	p.onMessageUnreliable = f
}

func (p *Peer) OnOpen(f func()) {
	p.onOpen = f
}

func (p *Peer) OnClose(f func()) {
	p.onClose = f
}

func (p *Peer) OnTrack(f func(*webrtc.TrackLocalStaticRTP)) {
	p.onTrack = f
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

func (p *Peer) wrap(sourcePeer *Peer, data []byte) ([]byte, error) {
	w := &Wrap{
		PeerID: sourcePeer.ID,
		Data:   data,
	}
	return msgpack.Marshal(w)
}
