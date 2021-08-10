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
		this.audio.volume = volume;
	}

	play(src: string): void {
		this.audio.src = src;
	}
}
