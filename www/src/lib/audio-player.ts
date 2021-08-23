// TODO: make this work on iOS. It requires for Audio.play() to be called after
// a user interaction on the page.
export default class AudioPlayer {
	private audio: HTMLAudioElement;

	constructor() {
		this.audio = new Audio();
		this.audio.autoplay = true;
		this.audio.controls = false;
		this.audio.loop = false;
		this.audio.muted = false;
		this.audio.addEventListener('canplay', () => this.audio.play());
	}

	setLoop(loop: boolean): void {
		this.audio.loop = loop;
	}

	setVolume(volume: number): void {
		if (volume < 0 || volume > 100) {
			return;
		}
		this.audio.volume = volume / 100;
	}

	play(src: string | undefined): void {
		this.audio.pause();
		this.audio.src = src || '';
	}
}
