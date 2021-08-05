import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import Signaling from '$lib/signaling';

export default class Peer {
	constructor(name, { onMouse, onPeer, onTrack }) {
		const onmessage = (event) => {
			const { peerID, data } = decode(event.data);

			if (data.MOUSE) {
				onMouse(peerID, data.MOUSE);
			}

			if (data.PEER) {
				console.log(`[onpeer] peer:${peerID} ${data.PEER.status}`);
				onPeer(peerID, data.PEER);
			}
		};

		/* Initialize signaling */

		this._signaling = new Signaling(`${env.OUTPOST_BASE_URL}/sessions/${encodeURIComponent(name)}`);

		/* Create and initialize WebRTC */

		this._peerConnection = new RTCPeerConnection({
			iceServers: [{ urls: env.STUN_SERVERS_URLS.split(',') }],
		});

		this._peerConnection.onconnectionstatechange = () =>
			console.log(`[webrtc] ${this._peerConnection.connectionState}`);

		this._peerConnection.onsignalingstatechange = () =>
			console.log(`[signal] ${this._peerConnection.signalingState}`);

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

		this._peerConnection.ontrack = ({ track, streams }) => {
			const stream = streams[0];
			const peerID = parseInt(stream.id, 10);
			console.log(`[ontrack] peer:${peerID} ${track.id}`);
			onTrack(peerID, track, stream);
		};

		/* WebRTC negotiation mecanism */

		this._peerConnection.onicecandidate = ({ candidate }) => this._signaling.send({ candidate });

		this._signaling.onmessage = async (data) => {
			if ('offer' in data) {
				await this._peerConnection.setRemoteDescription(data.offer);
				await this._peerConnection.setLocalDescription(await this._peerConnection.createAnswer());
				this._signaling.send({ answer: this._peerConnection.localDescription });
			} else if ('candidate' in data) {
				await this._peerConnection.addIceCandidate(data.candidate);
			}
		};

		this.negotiate();
	}

	negotiate() {
		return this._signaling.send({ negotiate: true });
	}

	addTracks(stream) {
		for (const track of stream.getTracks()) {
			this._peerConnection.addTrack(track, stream);
		}
		this.negotiate();
	}

	removeAudioTracks() {
		return this.removeTracksByKind('audio');
	}

	removeVideoTracks() {
		return this.removeTracksByKind('video');
	}

	removeTracksByKind(kind) {
		for (const sender of this._peerConnection.getSenders()) {
			if (sender.track?.kind === kind) {
				this._peerConnection.removeTrack(sender);
			}
		}
	}

	reliable(payload) {
		// TODO: buffer any message that couldn't be sent
		this._reliable.send(encode(payload));
	}

	unreliable(payload) {
		this._unreliable.send(encode(payload));
	}
}
