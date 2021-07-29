import { encode, decode } from '@msgpack/msgpack';
import * as env from '$lib/env';
import { IonSFUJSONRPCSignal } from 'ion-sdk-js/lib/signal/json-rpc-impl';

export default class Peer {
	constructor(roomName, IonClient, { onMouse, onPeer, onTrack }) {
		const onMessage = (event) => {
			const { peerID, data } = decode(event.data);

			if (data.MOUSE) {
				onMouse(peerID, data.MOUSE);
			}

			if (data.PEER) {
				onPeer(peerID, data.PEER);
			}
		};

		const signal = new IonSFUJSONRPCSignal(
			`${env.OUTPOST_BASE_URL}/rooms/${encodeURIComponent(roomName)}`,
		);

		this._client = new IonClient(signal, {
			iceServers: [{ urls: env.STUN_SERVERS_URLS.split(',') }],
		});

		signal.onopen = () => this._client.join(roomName);

		this._client.onconnectionstatechange = () => {
			console.log(`webrtc ${this._client.connectionState}`);

			if (this._client.connectionState === 'connected') {
				if (!this._reliable) {
					const reliable = this._client.createDataChannel('reliable', {
						negotiated: true,
						id: 0,
						//
						ordered: false,
					});
					reliable.onmessage = onMessage;
					reliable.onopen = () => {
						this._reliable = reliable;
					};
				}

				if (!this._unreliable) {
					const unreliable = this._client.createDataChannel('unreliable', {
						negotiated: true,
						id: 1,
						//
						ordered: false,
						maxRetransmits: 0,
					});
					unreliable.onmessage = onMessage;
					unreliable.onopen = () => {
						this._unreliable = unreliable;
					};
				}
			}
		};

		this._client.ontrack = (_track, stream) => onTrack(parseInt(stream.id, 10), stream);
	}

	publishStream(stream) {
		this._client.publish(stream);
	}

	sendReliable(payload) {
		this._reliable && this._reliable.send(encode(payload));
	}

	sendUnreliable(payload) {
		this._unreliable && this._unreliable.send(encode(payload));
	}
}
