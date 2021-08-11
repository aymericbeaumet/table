import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import { IonSFUJSONRPCSignal } from 'ion-sdk-js/lib/signal/json-rpc-impl.js';
import Entity from '$lib/entity';

export default class Peer {
	private client: any;
	private reliableDataChannel: any;
	private unreliableDataChannel: any;

	private publishedTimeout: any;
	private publishedStream: any;

	constructor(
		Client: any,
		tableName: string,
		{
			onBackgroundImageID,
			onBackgroundSoundID,
			onEntity,
			onSpeaker,
			onTrack,
			onTrackRemoved,
		}: {
			onBackgroundImageID: (resourceID: string) => void;
			onBackgroundSoundID: (resourceID: string) => void;
			onEntity: (streamID: string, entity: Entity) => void;
			onSpeaker: any;
			onTrack: any;
			onTrackRemoved: any;
		},
	) {
		const onmessage = (event: any) => {
			const { streamID, payload } = decode(event.data) as any;

			if (payload.BACKGROUND_IMAGE_ID) {
				return onBackgroundImageID(payload.BACKGROUND_IMAGE_ID);
			}

			if (payload.BACKGROUND_SOUND_ID) {
				return onBackgroundSoundID(payload.BACKGROUND_SOUND_ID);
			}

			if (payload.ENTITY) {
				const entity = new Entity(payload.ENTITY.kind, payload.ENTITY.props);
				entity.id = payload.ENTITY.id || streamID;
				return onEntity(streamID, entity);
			}
		};

		/* Initialize signaling */

		const signal = new IonSFUJSONRPCSignal(
			`${env.OUTPOST_BASE_URL}/tables/${encodeURIComponent(tableName)}`,
		);

		/* Create and initialize WebRTC */

		this.client = new Client(signal, {
			iceServers: [{ urls: env.STUN_SERVERS_URLS.split(',') }],
		});

		/* Listen for datachannels, speakers and tracks */

		this.client.ondatachannel = (event: any) => {
			const dc = event.channel;
			switch (dc.label) {
				case 'reliable':
					this.reliableDataChannel = dc;
					dc.onmessage = onmessage;
					return;
				case 'unreliable':
					this.unreliableDataChannel = dc;
					dc.onmessage = onmessage;
					return;
				default:
					throw new Error(`implementation error: unexpected channel "${dc.label}"`);
			}
		};

		this.client.onspeaker = onSpeaker;

		this.client.ontrack = (_track: any, stream: any) => {
			stream.onremovetrack = () => onTrackRemoved(stream);
			onTrack(stream);
		};

		/* Join when the signaling channel is ready */

		signal.onopen = () => this.client.join(name);
	}

	publish(stream: any): void {
		clearTimeout(this.publishedTimeout);

		if (this.publishedStream) {
			this.publishedStream.unpublish();
		}

		try {
			this.client.publish(stream);
			this.publishedStream = stream;
		} catch (error) {
			console.error(error);
			this.publishedTimeout = setTimeout(() => {
				this.publish(stream);
			}, 1000);
		}
	}

	reliable(payload: any): void {
		// TODO: buffer any message that couldn't be sent
		this.reliableDataChannel && this.reliableDataChannel.send(encode(payload));
	}

	unreliable(payload: any): void {
		this.unreliableDataChannel && this.unreliableDataChannel.send(encode(payload));
	}
}
