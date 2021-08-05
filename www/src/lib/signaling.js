export default class Signaling {
	constructor(url) {
		this._buffer = [];
		this._isOpen = false;
		this._url = url;
		this._nextFlush;

		this._init();
	}

	_init() {
		const ws = new WebSocket(this._url);

		ws.onerror = (error) => {
			console.error('websocket:', error);
		};

		ws.onopen = () => {
			this._isOpen = true;
		};

		ws.onmessage = (event) => {
			this.onmessage(JSON.parse(event.data));
		};

		ws.onclose = () => {
			this._isOpen = false;
		};

		this._isOpen = false;
		this._ws = ws;
	}

	send(message) {
		this.queue(message);
		this.flush();
	}

	queue(message) {
		this._buffer.push(JSON.stringify(message));
	}

	flush() {
		clearTimeout(this._nextFlush);

		if (!this._isOpen) {
			this._nextFlush = setTimeout(() => this.flush(), 1000);
			return;
		}

		while (this._buffer.length > 0) {
			const payload = this._buffer[0];
			try {
				this._ws.send(payload);
				this._buffer.shift();
			} catch (error) {
				this._nextFlush = setTimeout(() => this.flush(), 1000);
				console.error(error);
				return;
			}
		}
	}
}
