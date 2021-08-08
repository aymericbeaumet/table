import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import { IonSFUJSONRPCSignal } from 'ion-sdk-js/lib/signal/json-rpc-impl.js';
import Entity from '$lib/entity';

export default class Peer {
	constructor(
		Client,
		name,
		{ onBackgroundImage, onBackgroundSound, onEntity, onSpeaker, onTrack, onTrackRemoved },
	) {
		const onmessage = (event) => {
			const { streamID, payload } = decode(event.data);

			if (payload.BACKGROUND_IMAGE) {
				return onBackgroundImage(streamID, payload.BACKGROUND_IMAGE);
			}

			if (payload.BACKGROUND_SOUND) {
				return onBackgroundSound(streamID, payload.BACKGROUND_SOUND);
			}

			if (payload.ENTITY) {
				const entity = new Entity(payload.ENTITY.kind, payload.ENTITY.props);
				entity.id = payload.ENTITY.id || streamID;
				entity.key = `${entity.kind}:${entity.id}`;

				return onEntity(streamID, entity);
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
