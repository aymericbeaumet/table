package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pion/rtcp"
	webrtc "github.com/pion/webrtc/v3"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	rtcpPLIInterval = time.Second * 3
)

type Wrap struct {
	PeerID uint32             `msgpack:"peerID"`
	Data   msgpack.RawMessage `msgpack:"data"`
}

type Peer struct {
	ID     uint32 `msgpack:"id"`
	Status string `msgpack:"status"`

	peerConnection        *webrtc.PeerConnection
	signalDataChannel     *webrtc.DataChannel
	reliableDataChannel   *webrtc.DataChannel
	unreliableDataChannel *webrtc.DataChannel

	onOpen              func()
	onClose             func()
	onMessageReliable   func([]byte)
	onMessageUnreliable func([]byte)
	onTrack             func(webrtc.TrackLocal)
	onOffer             func(webrtc.SessionDescription)
}

func NewPeer(id uint32) (*Peer, error) {
	p := &Peer{
		ID:     id,
		Status: "offline",

		onOpen:              func() {},
		onClose:             func() {},
		onMessageReliable:   func([]byte) {},
		onMessageUnreliable: func([]byte) {},
		onTrack:             func(webrtc.TrackLocal) {},
		onOffer:             func(webrtc.SessionDescription) {},
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

	p.peerConnection.OnNegotiationNeeded(func() {
		offer, err := p.peerConnection.CreateOffer(nil)
		if err != nil {
			panic(err)
		}
		if err := p.peerConnection.SetLocalDescription(offer); err != nil {
			panic(err)
		}
		p.onOffer(offer)
	})

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

	// Listen for tracks

	if _, err = p.peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}

	p.peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for range ticker.C {
				if rtcpSendErr := p.peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}}); rtcpSendErr != nil {
					fmt.Println(rtcpSendErr)
				}
			}
		}()

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

	return p, nil
}

func (p *Peer) AddLocalTrack(track webrtc.TrackLocal) error {
	rtpSender, err := p.peerConnection.AddTrack(track)

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	return err
}

func (p *Peer) AddRemoteOffer(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	// Set remote description
	if err := p.peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	// Create answer
	answer, err := p.peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err = p.peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return p.peerConnection.LocalDescription(), nil
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

func (p *Peer) OnTrack(f func(webrtc.TrackLocal)) {
	p.onTrack = f
}

func (p *Peer) OnOffer(f func(webrtc.SessionDescription)) {
	p.onOffer = f
}

func (p *Peer) OnICECandidate(f func(candidate *webrtc.ICECandidate)) {
	p.peerConnection.OnICECandidate(f)
}

func (p *Peer) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return p.peerConnection.AddICECandidate(candidate)
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
