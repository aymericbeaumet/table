let audio;

export function play(src) {
	audio = audio || new Audio();
	audio.pause();

	audio.autoplay = true;
	audio.controls = false;
	audio.loop = false;
	audio.muted = false;
	audio.src = src;

	audio.play();
}
