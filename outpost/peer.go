package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pion/rtcp"
	webrtc "github.com/pion/webrtc/v3"
	"github.com/vmihailenco/msgpack/v5"
)

type Wrap struct {
	PeerID uint32             `msgpack:"peerID"`
	Data   msgpack.RawMessage `msgpack:"data"`
}

type Peer struct {
	ID uint32

	peerConnection        *webrtc.PeerConnection
	signalDataChannel     *webrtc.DataChannel
	reliableDataChannel   *webrtc.DataChannel
	unreliableDataChannel *webrtc.DataChannel

	tracks []webrtc.TrackLocal

	onConnected         func()
	onMessageReliable   func([]byte)
	onMessageUnreliable func([]byte)
	onTrack             func(webrtc.TrackLocal)

	mu sync.RWMutex
}

func NewPeer(id uint32) (*Peer, error) {
	p := &Peer{
		ID:                  id,
		onConnected:         func() {},
		onMessageReliable:   func([]byte) {},
		onMessageUnreliable: func([]byte) {},
		onTrack:             func(webrtc.TrackLocal) {},
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
	p.reliableDataChannel, err = p.peerConnection.CreateDataChannel("reliable", &webrtc.DataChannelInit{
		Negotiated: &rNegotiated,
		ID:         &rID,
		//
		Ordered: &rOrdered,
	})
	if err != nil {
		return nil, err
	}
	p.reliableDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		p.onMessageReliable(msg.Data)
	})

	uNegotiated, uID, uOrdered, uMaxRestransmits := true, uint16(1), false, uint16(0)
	p.unreliableDataChannel, err = p.peerConnection.CreateDataChannel("unreliable", &webrtc.DataChannelInit{
		Negotiated: &uNegotiated,
		ID:         &uID,
		//
		Ordered:        &uOrdered,
		MaxRetransmits: &uMaxRestransmits,
	})
	if err != nil {
		return nil, err
	}
	p.unreliableDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		p.onMessageUnreliable(msg.Data)
	})

	p.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateConnected {
			p.onConnected()
		}
	})

	p.peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := p.peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}}); err != nil {
					fmt.Println("webrtc error:", err)
					return
				}
			}
		}()

		track, err := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", strconv.FormatUint(uint64(p.ID), 10))
		if err != nil {
			fmt.Println("webrtc error:", err)
			return
		}

		p.mu.Lock()
		p.tracks = append(p.tracks, track)
		p.mu.Unlock()

		p.onTrack(track)

		rtpBuf := make([]byte, 1400)
		for {
			i, _, err := remoteTrack.Read(rtpBuf)
			if err != nil {
				fmt.Println("webrtc error:", err)
				return
			}
			if _, err = track.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				fmt.Println("webrtc error:", err)
				return
			}
		}
	})

	return p, nil
}

func (p *Peer) AddTrack(track webrtc.TrackLocal) error {
	rtpSender, err := p.peerConnection.AddTrack(track)
	if err != nil {
		return err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	return nil
}

func (p *Peer) EachTrack(f func(webrtc.TrackLocal)) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, t := range p.tracks {
		f(t)
	}
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

func (p *Peer) OnMessageReliable(f func(msg []byte)) {
	p.onMessageReliable = f
}

func (p *Peer) OnMessageUnreliable(f func(msg []byte)) {
	p.onMessageUnreliable = f
}

func (p *Peer) OnConnected(f func()) {
	p.onConnected = f
}

func (p *Peer) OnTrack(f func(webrtc.TrackLocal)) {
	p.onTrack = f
}

func (p *Peer) OnICECandidate(f func(candidate *webrtc.ICECandidate)) {
	p.peerConnection.OnICECandidate(f)
}

func (p *Peer) OnNegotiationNeeded(f func()) {
	p.peerConnection.OnNegotiationNeeded(f)
}

func (p *Peer) LocalDescription() *webrtc.SessionDescription {
	return p.peerConnection.LocalDescription()
}

func (p *Peer) Close() error {
	return p.peerConnection.Close()
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
