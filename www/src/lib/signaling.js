export default class Signaling {
	constructor(url) {
		this.onmessage = () => {};

		this._buffer = [];
		this._url = url;
		this._ws = this._init();
	}

	_init() {
		const ws = new WebSocket(this._url);
		ws.onmessage = (event) => this.onmessage(JSON.parse(event.data));
		return ws;
	}

	send(payload) {
		try {
			this._ws.send(JSON.stringify(payload));
		} catch (error) {
			// TODO: buffer payload and try again ASAP
			console.error(error);
		}
	}
}
