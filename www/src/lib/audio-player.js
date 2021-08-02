export default class AudioPlayer {
	constructor() {
		this._audio = new Audio();
		this._audio.autoplay = true;
		this._audio.controls = false;
		this._audio.loop = false;
		this._audio.muted = false;
	}

	setLoop(loop) {
		this._audio.loop = loop;
	}

	setVolume(volume) {
		this._audio.volume = volume;
	}

	play(src) {
		this._audio.src = src;
	}
}
