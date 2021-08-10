export default class {
	private currentAudioInput?: string;
	private currentVideoInput?: string;

	constructor(private LocalStream: any, private framerate: number) {}

	async reset(
		desiredAudioInput: string,
		desiredVideoInput: string,
	): Promise<[any, boolean, boolean] | null> {
		// Idempotent
		const hasAudioChanged = desiredAudioInput !== this.currentAudioInput;
		const hasVideoChanged = desiredVideoInput !== this.currentVideoInput;
		if (!hasAudioChanged && !hasVideoChanged) {
			return null;
		}

		// Get user media
		const stream = await this.LocalStream.getUserMedia({
			audio: {
				deviceId: desiredAudioInput,
				autoGainControl: true,
				echoCancellation: true,
				noiseSuppression: true,
			},
			video: {
				deviceId: desiredVideoInput,
				width: { ideal: 360, max: 360 },
				height: { ideal: 360, max: 360 },
				frameRate: { ideal: this.framerate, max: this.framerate },
			},
			codec: 'vp8',
			resolution: 'qvga',
			sendEmptyOnMute: true,
			simulcast: true,
		});

		// Save new current audio/video inputs
		this.currentAudioInput = desiredAudioInput;
		this.currentVideoInput = desiredVideoInput;

		return [stream, hasAudioChanged, hasVideoChanged];
	}
}
