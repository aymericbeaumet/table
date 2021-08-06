import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import { IonSFUJSONRPCSignal } from 'ion-sdk-js/lib/signal/json-rpc-impl.js';

export default class Peer {
	constructor(Client, name, { onMouse, onSpeaker, onTrack, onTrackRemoved }) {
		const onmessage = (event) => {
			const { streamID, payload } = decode(event.data);
			console.log({ streamID, payload });

			if (payload.MOUSE) {
				onMouse(streamID, payload.MOUSE);
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

		/* Listen for datachannels, speakers and tracks */

		this._client.ondatachannel = (event) => {
			const dc = event.channel;
			switch (dc.label) {
				case 'reliable':
					this._reliable = dc;
					dc.onmessage = onmessage;
					return;
				case 'unreliable':
					this._unreliable = dc;
					dc.onmessage = onmessage;
					return;
				default:
					throw new Error(`implementation error: unexpected channel "${dc.label}"`);
			}
		};

		this._client.onspeaker = onSpeaker;

		this._client.ontrack = (_track, stream) => {
			stream.onremovetrack = () => onTrackRemoved(stream);
			onTrack(stream);
		};

		/* Join when the signaling channel is ready */

		signal.onopen = () => this._client.join(name);
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
