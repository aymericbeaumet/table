import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import { IonSFUJSONRPCSignal } from 'ion-sdk-js/lib/signal/json-rpc-impl.js';
import Entity from '$lib/entity';
import type { Client as ClientType, LocalStream as LocalStreamType } from 'ion-sdk-js';

export default class Peer {
	private client: ClientType;
	private reliableDataChannel: any;
	private unreliableDataChannel: any;

	private publishedTimeout?: ReturnType<typeof setTimeout>;
	private publishedStream?: any;

	constructor(
		Client: typeof ClientType,
		signalURL: string,
		tableName: string,
		{
			onAmbianceID,
			onAmbianceSoundVolume,
			onAmbianceEffectsVolume,
			onEntity,
			onScreen,
			onSpeaker,
			onTrack,
			onTrackRemoved,
		}: {
			onAmbianceID: (streamID: string, id: string) => void;
			onAmbianceSoundVolume: (streamID: string, volume: number) => void;
			onAmbianceEffectsVolume: (streamID: string, volume: number) => void;
			onEntity: (streamID: string, entity: Entity) => void;
			onScreen: (streamID: string, width: number, height: number) => void;
			onSpeaker: (streamIDs: Array<string>) => void;
			onTrack: (stream: LocalStreamType) => void;
			onTrackRemoved: (stream: LocalStreamType) => void;
		},
	) {
		const onmessage = (event: any) => {
			const { streamID, payload } = decode(event.data) as any;

			if (payload.AMBIANCE_ID) {
				return onAmbianceID(streamID, payload.AMBIANCE_ID);
			}

			if (payload.AMBIANCE_SOUND_VOLUME) {
				return onAmbianceSoundVolume(streamID, payload.AMBIANCE_SOUND_VOLUME);
			}

			if (payload.AMBIANCE_EFFECTS_VOLUME) {
				return onAmbianceEffectsVolume(streamID, payload.AMBIANCE_EFFECTS_VOLUME);
			}

			if (payload.ENTITY) {
				const entity = new Entity(payload.ENTITY.kind, payload.ENTITY.props);
				entity.id = payload.ENTITY.id || streamID;
				return onEntity(streamID, entity);
			}

			if (payload.SCREEN) {
				return onScreen(streamID, payload.SCREEN.width, payload.SCREEN.height);
			}
		};

		/* Initialize signaling */

		const signal = new IonSFUJSONRPCSignal(signalURL);

		/* Create and initialize WebRTC */

		this.client = new Client(signal, {
			iceServers: [{ urls: env.STUN_SERVERS_URLS.split(',') }],
			codec: 'vp8',
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

		signal.onopen = () => this.client.join(tableName, '');
	}

	publish(stream: LocalStreamType): void {
		if (this.publishedTimeout) {
			clearTimeout(this.publishedTimeout);
			this.publishedTimeout = undefined;
		}

		if (this.publishedStream) {
			this.publishedStream.unpublish();
			this.publishedStream = undefined;
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
		// TODO: buffer messages that couldn't be sent
		this.reliableDataChannel && this.reliableDataChannel.send(encode(payload));
	}

	unreliable(payload: any): void {
		this.unreliableDataChannel && this.unreliableDataChannel.send(encode(payload));
	}
}
