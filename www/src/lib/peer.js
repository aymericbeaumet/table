import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';

export default class Peer {
	constructor(ws, { onMouse, onPeer, onTrack }) {
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
			iceServers: [{ urls: env.STUN_SERVERS_URLS.split(',') }],
		});

		this._peerConnection.onconnectionstatechange = () => {
			console.log(`webrtc ${this._peerConnection.connectionState}`);
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

		/* Listen for tracks */

		this._peerConnection.ontrack = (event) => {
			const stream = event.streams[0];
			const peerID = parseInt(stream.id, 10);
			onTrack(peerID, event.track, stream);
		};

		/* WebRTC negotiation mecanism */

		let isNegotiationNeeded = false;
		let isOfferPending = false;

		this._peerConnection.onicecandidate = ({ candidate }) => ws.send(JSON.stringify({ candidate }));

		this._peerConnection.onnegotiationneeded = async () => {
			if (isOfferPending) {
				return;
			}
			await this._peerConnection.setLocalDescription();
			ws.send(JSON.stringify({ description: this._peerConnection.localDescription }));
			isNegotiationNeeded = false;
			isOfferPending = true;
		};

		ws.onmessage = async (event) => {
			const { negotiationneeded, description, candidate } = JSON.parse(event.data);

			if (negotiationneeded) {
				isNegotiationNeeded = true;
				this._peerConnection.onnegotiationneeded();
			} else if (description) {
				await this._peerConnection.setRemoteDescription(description);
				isOfferPending = false;
				if (isNegotiationNeeded) {
					await this._peerConnection.onnegotiationneeded();
				}
			} else if (candidate) {
				await this._peerConnection.addIceCandidate(candidate);
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
