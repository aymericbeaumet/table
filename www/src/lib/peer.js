import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import Signaling from '$lib/signaling';
import { debounce } from 'lodash';

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

		/* Debounced onnegotiationneeded handler */

		this._onnegotiationneeded = debounce(async () => {
			if (this._isOfferPending) {
				this._isNegotiationNeeded = true;
				return;
			}
			this._isNegotiationNeeded = false;
			this._isOfferPending = true;
			const offer = await this._peerConnection.createOffer();
			await this._peerConnection.setLocalDescription(offer);
			this._signaling.send({ description: this._peerConnection.localDescription });
		}, 250);

		/* Initialize signaling */

		this._signaling = new Signaling(`${env.OUTPOST_BASE_URL}/rooms/${encodeURIComponent(name)}`);

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

		this._onnegotiationneeded();

		/* Listen for tracks */

		this._peerConnection.ontrack = ({ track, streams }) => {
			const stream = streams[0];
			const peerID = parseInt(stream.id, 10);
			console.log(`[ontrack] peer:${peerID} ${track.id}`);
			onTrack(peerID, track, stream);
		};

		/* WebRTC negotiation mecanism */

		this._isNegotiationNeeded = false;
		this._isOfferPending = false;

		this._peerConnection.onicecandidate = ({ candidate }) => this._signaling.send({ candidate });

		this._signaling.onmessage = async (data) => {
			if ('negotiationneeded' in data) {
				this._onnegotiationneeded();
			} else if ('description' in data) {
				await this._peerConnection.setRemoteDescription(data.description);
				this._isOfferPending = false;
				if (this._isNegotiationNeeded) {
					await this._onnegotiationneeded();
				}
			} else if ('candidate' in data) {
				await this._peerConnection.addIceCandidate(data.candidate);
			}
		};
	}

	addTracks(stream) {
		for (const track of stream.getTracks()) {
			this._peerConnection.addTrack(track, stream);
		}
		this._onnegotiationneeded();
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
