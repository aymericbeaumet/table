import type { LocalStream as LocalStreamType } from 'ion-sdk-js';

export default class {
	private currentAudioInput?: string;
	private currentVideoInput?: string;

	constructor(private LocalStream: typeof LocalStreamType, private framerate: number) {}

	async reset(
		desiredAudioInput: string,
		desiredVideoInput: string,
	): Promise<[LocalStreamType, boolean, boolean] | null> {
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

	static async detect(): Promise<[string, string]> {
		let audioInputDeviceID = '';
		let videoInputDeviceID = '';

		const devices = await navigator.mediaDevices.enumerateDevices();
		for (const device of devices) {
			const isDefault = device.deviceId === 'default';
			const isBuiltin = device.label.toLowerCase().includes('facetime');
			const isFound =
				(device.kind === 'audioinput' && audioInputDeviceID) ||
				(device.kind === 'videoinput' && videoInputDeviceID);

			if (isDefault || (isBuiltin && !isFound)) {
				if (device.kind === 'audioinput') {
					audioInputDeviceID = device.deviceId;
				} else if (device.kind === 'videoinput') {
					videoInputDeviceID = device.deviceId;
				}
			}
		}

		return [audioInputDeviceID, videoInputDeviceID];
	}
}
