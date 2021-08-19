import getVideoId from 'get-video-id';

export default class Entity {
	public kind: string;
	public id?: string;
	public props: any;

	constructor(kind: string, props = {}) {
		this.kind = kind;
		this.id = undefined;
		this.props = props;
	}

	get key(): string {
		return `${this.kind}:${this.id}`;
	}

	toString(): string {
		switch (this.kind) {
			case 'iframe':
				return this.props.src;
			case 'text':
				return this.props.value;
			default:
				throw new Error(`implementation error: unexpected kind "${this.kind}"`);
		}
	}

	static infer(value: string): Array<Entity> {
		const out = new Array<Entity>();

		if (value.length === 0) {
			return out;
		}

		const { service, id } = getVideoId(value);
		if (id) {
			switch (service) {
				case 'vimeo':
					out.push(
						new Entity('iframe', {
							icon: 'vimeo',
							src: `https://player.vimeo.com/video/${encodeURIComponent(id)}`,
						}),
					);
					break;
				case 'youtube':
					out.push(
						new Entity('iframe', {
							icon: 'youtube',
							src: `https://www.youtube-nocookie.com/embed/${encodeURIComponent(id)}`,
						}),
					);
					break;
			}
		}

		{
			const soundcloudRegex = /https?:\/\/soundcloud\.com\/[a-zA-Z0-9-]+\/[a-zA-Z0-9-]+/;
			const soundcloudMatches = soundcloudRegex.exec(value);
			if (soundcloudMatches && soundcloudMatches[0]) {
				out.push(
					new Entity('iframe', {
						icon: 'soundcloud',
						src: `https://w.soundcloud.com/player/?url=${encodeURIComponent(
							soundcloudMatches[0],
						)}&auto_play=false&show_artwork=true`,
					}),
				);
			}
		}

		{
			const spotifyRegex = /https?:\/\/open\.spotify\.com\/([a-z]+)\/([A-Za-z0-9]+)\/?/;
			const spotifyMatches = spotifyRegex.exec(value);
			if (spotifyMatches && spotifyMatches[1] && spotifyMatches[2]) {
				out.push(
					new Entity('iframe', {
						icon: 'spotify',
						src: `https://open.spotify.com/embed/${encodeURIComponent(
							spotifyMatches[1],
						)}/${encodeURIComponent(spotifyMatches[2])}`,
					}),
				);
			}
		}

		{
			const twitchRegex = /https?:\/\/(?:www\.)?twitch\.tv\/([a-z0-9]+)/;
			const twitchMatches = twitchRegex.exec(value);
			if (twitchMatches && twitchMatches[1]) {
				out.push(
					new Entity('iframe', {
						icon: 'twitch',
						src: `https://player.twitch.tv/?channel=${encodeURIComponent(
							twitchMatches[1],
						)}&parent=${encodeURIComponent(window.location.hostname)}`,
					}),
				);
			}
		}

		out.push(
			new Entity('text', {
				value,
			}),
		);

		return out;
	}
}
