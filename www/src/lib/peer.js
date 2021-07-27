import { encode, decode } from '@msgpack/msgpack';

export default class Peer {
	constructor(signaling, { onMouse, onPeer, onTrack }) {
		const onmessage = (event) => {
			const { peerID, data } = decode(event.data);

			if (data.MOUSE) {
				onMouse(peerID, data.MOUSE);
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
			signaling.send(JSON.stringify({ DESCRIPTION: offer }));
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

		this._peerConnection.addTransceiver('video');
		this._peerConnection.ontrack = onTrack;

		/* Deal with signaling messages */

		signaling.onmessage = async (event) => {
			const payload = JSON.parse(event.data);

			if (payload.DESCRIPTION) {
				if (payload.DESCRIPTION.type !== 'answer') {
					throw new Error(`unexpected session description type "${payload.DESCRIPTION.type}"`);
				}
				this._peerConnection.setRemoteDescription(new RTCSessionDescription(payload.DESCRIPTION));
			}

			if (payload.ICE_CANDIDATE) {
				this._peerConnection.addIceCandidate(payload.ICE_CANDIDATE);
			}

			if (payload.NEGOTIATION_NEEDED) {
				this._peerConnection.onnegotiationneeded();
			}
		};
	}

	addLocalStream(stream) {
		for (const track of stream.getTracks()) {
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
