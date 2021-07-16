import ReconnectingWebSocket from 'reconnecting-websocket';
import SimplePeer from 'simple-peer';

const WEBSOCKET_BASE_URL = process.env.WEBSOCKET_BASE_URL || 'ws://localhost:8080';

export default class Network {
	constructor(name) {
		this.name = name;
		this.peers = {};
	}

	start() {
		const url = `${WEBSOCKET_BASE_URL}/rooms/${encodeURIComponent(this.name)}`;
		const ws = new ReconnectingWebSocket(url);

		ws.onmessage = (event) => {
			const { PEER_STATUS, PEER_SIGNAL } = JSON.parse(event.data);

			// Update the peer connection following a status update
			if (PEER_STATUS) {
				const { peerID, status } = PEER_STATUS;
				switch (status) {
					case 'online': {
						const peer = this.peers[peerID] || newPeer(ws, peerID, { initiator: true });
						this.peers[peerID] = peer;
						break;
					}
					case 'offline': {
						const peer = this.peers[peerID];
						if (peer) {
							peer.destroy();
							delete this.peers[peerID];
						}
						break;
					}
				}
			}

			// Accept a signal (offer/answer) from another user
			if (PEER_SIGNAL) {
				const { peerID, signal } = PEER_SIGNAL;
				const peer = this.peers[peerID] || newPeer(ws, peerID, { initiator: false });
				peer.signal(signal);
				this.peers[peerID] = peer;
			}
		};
	}
}

function newPeer(ws, peerID, { initiator, stream }) {
	console.log(`[${peerID}] New peer`);
	const peer = new SimplePeer({
		trickle: false,
		initiator,
		stream
	});

	peer.on('error', (err) => {
		console.error(`[${peerID}] error:`, err);
	});

	peer.once('signal', (data) => {
		ws.send(
			JSON.stringify({
				SIGNAL: { peerID, signal: data }
			})
		);
	});

	peer.on('data', (data) => {
		console.log(peerID, JSON.parse(data));
	});

	peer.on('stream', (stream) => {
		console.log(`[${peerID}] Incoming stream`);

		const cancel = newVideoDOM(stream);
		peer.once('close', cancel);
	});

	return peer;
}
