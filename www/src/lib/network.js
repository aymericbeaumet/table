import { encode, decode } from '@msgpack/msgpack';

export default class Network {
	constructor(roomName, stream, { onPeer, onMouseMove }) {
		this._roomName = roomName;
		this._stream = stream;
		this._seq = 0;

		const onmessage = (event) => {
			const { peerID, data } = decode(event.data);

			// TODO: drop the outdated message by leveraging the __seq__ number by message key

			if (data.PEER) {
				onPeer(peerID, data.PEER);
			}

			if (data.MOUSE_MOVE) {
				onMouseMove(peerID, data.MOUSE_MOVE);
			}
		};

		/* Prepare WebRTC */

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

		/* Prepare WebRTC unordered but reliable data channel */

		this._reliable = this._peerConnection.createDataChannel('reliable', {
			ordered: false,
		});

		this._reliable.onmessage = onmessage;

		/* Prepare WebRTC unordered and unreliable data channel */

		this._unreliable = this._peerConnection.createDataChannel('unreliable', {
			ordered: false,
			maxRetransmits: 0,
		});

		this._unreliable.onmessage = onmessage;
	}

	async init() {
		for (const track of this._stream.getTracks()) {
			this._peerConnection.addTrack(track, this._stream);
		}

		const offer = await this._peerConnection.createOffer();

		this._peerConnection.setLocalDescription(offer);

		const res = await fetch(`${import.meta.env.VITE_API_BASE_URL}/rooms`, {
			method: 'POST',
			body: JSON.stringify({ name: this._roomName, offer }),
		});
		if (res.status !== 200) {
			throw new Error(`unexpected status code ${res.status}`);
		}
		const { answer } = await res.json();

		this._peerConnection.setRemoteDescription(new RTCSessionDescription(answer));
	}

	reliable(payload) {
		payload.__seq__ = this._seq++;
		this._reliable.send(encode(payload));
	}

	unreliable(payload) {
		payload.__seq__ = this._seq++;
		this._unreliable.send(encode(payload));
	}
}
