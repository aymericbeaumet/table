export default class AudioPlayer {
	private audio: HTMLAudioElement;

	constructor() {
		this.audio = new Audio();
		this.audio.autoplay = true;
		this.audio.controls = false;
		this.audio.loop = false;
		this.audio.muted = false;
	}

	setLoop(loop: boolean): void {
		this.audio.loop = loop;
	}

	setVolume(volume: number): void {
		if (volume < 0 || volume > 100) {
			throw new Error(`implementation error: volume should be between 0 and 100, got ${volume}`);
		}
		this.audio.volume = volume / 100;
	}

	play(src: string): void {
		this.audio.src = src;
	}
}
