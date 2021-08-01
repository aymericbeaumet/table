package sfu

import webrtc "github.com/pion/webrtc/v3"

type peerEvent struct {
	connectionOpen    *peerEventConnectionOpen
	connectionClose   *peerEventConnectionClose
	reliableMessage   *peerEventReliableMessage
	unreliableMessage *peerEventUnreliableMessage
	remoteTrack       *peerEventRemoteTrack
}

type peerEventConnectionOpen struct{}

type peerEventConnectionClose struct{}

type peerEventReliableMessage struct {
	data []byte
}

type peerEventUnreliableMessage struct {
	data []byte
}

type peerEventRemoteTrack struct {
	remoteTrack *webrtc.TrackRemote
}
