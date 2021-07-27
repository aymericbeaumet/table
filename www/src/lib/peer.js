import { encode, decode } from '@msgpack/msgpack';

export default class Peer {
	constructor(ws, { onMouse, onPeer, onTrack }) {
		const onmessage = (event) => {
			const { peerID, data } = decode(event.data);

			if (data.MOUSE) {
				onMouse(peerID, data.MOUSE);
			}

			if (data.PEER) {
				onPeer(peerID, data.PEER);
			}
		};

		/* Create and initialize WebRTC */

		this._peerConnection = new RTCPeerConnection({
			iceServers: [{ urls: import.meta.env.VITE_STUN_SERVERS_URLS.split(',') }],
		});

		this._peerConnection.onconnectionstatechange = () => {
			console.log(`webrtc ${this._peerConnection.connectionState}`);
		};

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

		this._peerConnection.ontrack = onTrack;

		/*
		 * https://w3c.github.io/webrtc-pc/#perfect-negotiation-example
		 * The perfect negotiation logic, separated from the rest of the application
		 */

		const polite = true;

		// keep track of some negotiation state to prevent races and errors
		let makingOffer = false;
		let ignoreOffer = false;
		let isSettingRemoteAnswerPending = false;

		// send any ice candidates to the other peer
		this._peerConnection.onicecandidate = ({ candidate }) => ws.send(JSON.stringify({ candidate }));

		// let the "negotiationneeded" event trigger offer generation
		this._peerConnection.onnegotiationneeded = async () => {
			try {
				makingOffer = true;
				await this._peerConnection.setLocalDescription();
				ws.send(JSON.stringify({ description: this._peerConnection.localDescription }));
			} catch (err) {
				console.error(err);
			} finally {
				makingOffer = false;
			}
		};

		ws.onmessage = async (event) => {
			const { negotiationneeded, description, candidate } = JSON.parse(event.data);
			try {
				if (negotiationneeded) {
					this._peerConnection.onnegotiationneeded();
				} else if (description) {
					// An offer may come in while we are busy processing SRD(answer).
					// In this case, we will be in "stable" by the time the offer is processed
					// so it is safe to chain it on our Operations Chain now.
					const readyForOffer =
						!makingOffer &&
						(this._peerConnection.signalingState === 'stable' || isSettingRemoteAnswerPending);
					const offerCollision = description.type === 'offer' && !readyForOffer;

					ignoreOffer = !polite && offerCollision;
					if (ignoreOffer) {
						return;
					}
					isSettingRemoteAnswerPending = description.type === 'answer';
					await this._peerConnection.setRemoteDescription(description); // SRD rolls back as needed
					isSettingRemoteAnswerPending = false;
					if (description.type == 'offer') {
						await this._peerConnection.setLocalDescription();
						ws.send(JSON.stringify({ description: this._peerConnection.localDescription }));
					}
				} else if (candidate) {
					try {
						await this._peerConnection.addIceCandidate(candidate);
					} catch (err) {
						if (!ignoreOffer) throw err; // Suppress ignored offer's candidates
					}
				}
			} catch (err) {
				console.error(err);
			}
		};
	}

	addLocalStream(stream) {
		for (const track of stream.getTracks()) {
			this._peerConnection.addTrack(track, stream);
		}
	}

	reliable(payload) {
		this._reliable.send(encode(payload));
	}

	unreliable(payload) {
		this._unreliable.send(encode(payload));
	}
}
