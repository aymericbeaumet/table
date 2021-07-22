export default class Network {
	constructor(roomName, stream) {
		this.roomName = roomName;
		this.stream = stream;

		/* Prepare WebRTC */

		this.pc = new RTCPeerConnection({
			iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
		});

		this.pc.onconnectionstatechange = () => {
			console.log(`webrtc ${this.pc.connectionState}`);
		};

		/* Prepare WebRTC unordered but reliable data channel */

		const reliable = this.pc.createDataChannel('reliable', {
			ordered: false,
		});

		reliable.onopen = () => {
			console.log('webrtc reliable channel open');
		};

		reliable.onmessage = (event) => {
			console.log('webrtc reliable channel message', event.data);
		};

		/* Prepare WebRTC unordered and unreliable data channel */

		const unreliable = this.pc.createDataChannel('unreliable', {
			ordered: false,
			maxRetransmits: 0,
		});

		unreliable.onopen = () => {
			console.log('webrtc unreliable channel open');
		};

		unreliable.onmessage = (event) => {
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
}
