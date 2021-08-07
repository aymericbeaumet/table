import getVideoId from 'get-video-id';
import EntityCursor from '$lib/components/EntityCursor.svelte';
import EntityIframe from '$lib/components/EntityIframe.svelte';
import EntityText from '$lib/components/EntityText.svelte';

export default class Entity {
	constructor(kind, props = {}) {
		this.kind = kind;
		this.id = undefined;
		this.props = props;
	}

	asComponentProps() {
		switch (this.kind) {
			case 'cursor':
				return [EntityCursor, this.props];
			case 'iframe':
				return [EntityIframe, this.props];
			case 'text':
				return [EntityText, this.props];
			default:
				throw new Error(`implementation error: unexpected entity kind "${this.kind}"`);
		}
	}

	static infer(value) {
		const out = [];

		if (value.length === 0) {
			return out;
		}

		const { service, id } = getVideoId(value);
		switch (service) {
			case 'vimeo':
				out.push(
					new Entity('iframe', {
						src: `https://player.vimeo.com/video/${encodeURIComponent(id)}`,
					}),
				);
				break;
			case 'youtube':
				out.push(
					new Entity('iframe', {
						src: `https://www.youtube-nocookie.com/embed/${encodeURIComponent(id)}`,
					}),
				);
				break;
		}

		{
			const matches = /https?:\/\/open\.spotify\.com\/([a-z]+)\/([A-Za-z0-9]+)\/?/.exec(value);
			if (matches && matches.length === 3) {
				out.push(
					new Entity('iframe', {
						src: `https://open.spotify.com/embed/${encodeURIComponent(
							matches[1],
						)}/${encodeURIComponent(matches[2])}`,
					}),
				);
			}
		}

		{
			const matches = /https?:\/\/(?:www\.)?twitch\.tv\/([a-z0-9]+)/.exec(value);
			if (matches && matches.length === 2) {
				out.push(
					new Entity('iframe', {
						src: `https://player.twitch.tv/?channel=${encodeURIComponent(
							matches[1],
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
