import { encode, decode } from '@msgpack/msgpack';

export default class Peer {
	constructor(signaling, { onMouseDown, onMouseMove, onMouseUp, onPeer, onTrack }) {
		const onmessage = (event) => {
			const { peerID, data } = decode(event.data);

			if (data.MOUSE_DOWN) {
				onMouseDown(peerID, data.MOUSE_DOWN);
			}

			if (data.MOUSE_MOVE) {
				onMouseMove(peerID, data.MOUSE_MOVE);
			}

			if (data.MOUSE_UP) {
				onMouseUp(peerID, data.MOUSE_UP);
			}

			if (data.PEER) {
				onPeer(peerID, data.PEER);
			}
		};

		/* Create and initialize WebRTC */

		this._peerConnection = new RTCPeerConnection({
			iceServers: [
				{
					urls: import.meta.env.VITE_STUN_SERVERS_URLS.split(','),
				},
			],
		});

		this._peerConnection.onconnectionstatechange = () => {
			console.log(`webrtc ${this._peerConnection.connectionState}`);
		};

		this._peerConnection.onnegotiationneeded = async () => {
			const offer = await this._peerConnection.createOffer();
			await this._peerConnection.setLocalDescription(offer);
			signaling.send(JSON.stringify({ OFFER: offer }));
		};

		this._peerConnection.onicecandidate = (event) => {
			signaling.send(JSON.stringify({ ICE_CANDIDATE: event.candidate }));
		};

		/* Create pre-negotiated data channels */

		this._reliable = this._peerConnection.createDataChannel('reliable', {
			negotiated: true,
			id: 0,
			//
			ordered: false,
		});
		this._reliable.onmessage = onmessage;

		this._unreliable = this._peerConnection.createDataChannel('unreliable', {
			negotiated: true,
			id: 1,
			//
			ordered: false,
			maxRetransmits: 0,
		});
		this._unreliable.onmessage = onmessage;

		/* Prepare tracks */

		this._localTracksIDs = new Set();
		this._peerConnection.addTransceiver('video');
		this._peerConnection.ontrack = (event) => {
			// Do not forwards our local tracks
			if (this._localTracksIDs.has(event.transceiver.sender.track.id)) {
				return;
			}
			onTrack(event);
		};

		/* Deal with signaling messages */

		signaling.onmessage = (event) => {
			const payload = JSON.parse(event.data);

			if (payload.ANSWER) {
				this._peerConnection.setRemoteDescription(new RTCSessionDescription(payload.ANSWER));
			}

			if (payload.CANDIDATE) {
				this._peerConnection.addIceCandidate(payload.CANDIDATE);
			}
		};
	}

	addLocalStream(stream) {
		for (const track of stream.getTracks()) {
			this._localTracksIDs.add(track.id);
			this._peerConnection.addTrack(track, stream);
		}
	}

	reliable(payload) {
		this._reliable.send(encode(payload));
	}

	unreliable(payload) {
		this._unreliable.send(encode(payload));
	}
}
