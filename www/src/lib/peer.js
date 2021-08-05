import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import { IonSFUJSONRPCSignal } from 'ion-sdk-js/lib/signal/json-rpc-impl.js';

export default class Peer {
	constructor(Client, name, { onMouse, onPeer, onTrack, onTrackRemoved }) {
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

		const signal = new IonSFUJSONRPCSignal(
			`${env.OUTPOST_BASE_URL}/sessions/${encodeURIComponent(name)}`,
		);

		/* Create and initialize WebRTC */

		this._client = new Client(signal, {
			iceServers: [{ urls: env.STUN_SERVERS_URLS.split(',') }],
		});

		signal.onopen = async () => await this._client.join(name);

		/* Create pre-negotiated data channels */

		//this._reliable = this._client.createDataChannel('reliable', {
		//negotiated: true,
		//id: 0,
		////
		//ordered: false,
		//});
		//this._reliable.onmessage = onmessage;

		//this._unreliable = this._client.createDataChannel('unreliable', {
		//negotiated: true,
		//id: 1,
		////
		//ordered: false,
		//maxRetransmits: 0,
		//});
		//this._unreliable.onmessage = onmessage;

		/* Listen for tracks */

		this._client.ontrack = (track, stream) => {
			stream.onremovetrack = (track) => onTrackRemoved('', track, stream);
			onTrack('', track, stream);
		};
	}

	publish(stream) {
		if (this._publishedStream) {
			this._publishedStream.unpublish();
			this._publishedStream = null;
		}

		try {
			this._client.publish(stream);
			this._pendingStream = null;
			this._publishedStream = stream;
		} catch (error) {
			this._pendingStream = stream;
		}
	}

	reliable(payload) {
		// TODO: buffer any message that couldn't be sent
		this._reliable && this._reliable.send(encode(payload));
	}

	unreliable(payload) {
		this._unreliable && this._unreliable.send(encode(payload));
	}
}
