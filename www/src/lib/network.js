export default class Network {
	constructor(roomName, stream) {
		/* Prepare WebRTC */

		const pc = new RTCPeerConnection({
			iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
		});

		pc.onconnectionstatechange = () => {
			console.log(`webrtc ${pc.connectionState}`);
		};

		stream.getTracks().forEach((track) => pc.addTrack(track, stream));

		/* Prepare WebSocket */

		const ws = new WebSocket(
			`${import.meta.env.VITE_WEBSOCKET_BASE_URL}/rooms/${encodeURIComponent(roomName)}`,
		);

		ws.onopen = async () => {
			console.log('websocket open');
			const offer = await pc.createOffer();
			pc.setLocalDescription(offer);
			ws.send(JSON.stringify({ OFFER: offer }));
		};

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			console.log('websocket message', data);

			const { ANSWER } = data;
			if (ANSWER) {
				pc.setRemoteDescription(new RTCSessionDescription(ANSWER));
			}
		};
	}
}
