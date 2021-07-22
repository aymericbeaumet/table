export default class Network {
	constructor(roomName, stream) {
		this.roomName = roomName;
		this.stream = stream;
		this.seq = 0;

		/* Prepare WebRTC */

		this.pc = new RTCPeerConnection({
			iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
		});

		this.pc.onconnectionstatechange = () => {
			console.log(`webrtc ${this.pc.connectionState}`);
		};

		/* Prepare WebRTC unordered but reliable data channel */

		this.reliable = this.pc.createDataChannel('reliable', {
			ordered: false,
		});

		this.reliable.onopen = () => {
			console.log('webrtc reliable channel open');
		};

		this.reliable.onmessage = (event) => {
			console.log('webrtc reliable channel message', event.data);
		};

		/* Prepare WebRTC unordered and unreliable data channel */

		this.unreliable = this.pc.createDataChannel('unreliable', {
			ordered: false,
			maxRetransmits: 0,
		});

		this.unreliable.onopen = () => {
			console.log('webrtc unreliable channel open');
		};

		this.unreliable.onmessage = (event) => {
			console.log('webrtc unreliable channel message', event.data);
		};
	}

	async init() {
		for (const track of this.stream.getTracks()) {
			this.pc.addTrack(track, this.stream);
		}

		const offer = await this.pc.createOffer();

		this.pc.setLocalDescription(offer);

		const res = await fetch(`${import.meta.env.VITE_API_BASE_URL}/rooms`, {
			method: 'POST',
			body: JSON.stringify({ name: this.roomName, offer }),
		});
		const { answer } = await res.json();

		this.pc.setRemoteDescription(new RTCSessionDescription(answer));
	}

	send(payload, { reliable }) {
		payload.__seq__ = this.seq++;
		const json = JSON.stringify(payload);
		if (reliable) {
			this.reliable.send(json);
		} else {
			this.unreliable.send(json);
		}
	}
}
