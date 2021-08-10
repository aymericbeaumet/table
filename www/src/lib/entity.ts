import getVideoId from 'get-video-id';
import EntityCursor from '$lib/components/EntityCursor.svelte';
import EntityIframe from '$lib/components/EntityIframe.svelte';
import EntityText from '$lib/components/EntityText.svelte';

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

	asComponentProps(): [any, any] {
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
		}

		{
			const sMatches = /https?:\/\/open\.spotify\.com\/([a-z]+)\/([A-Za-z0-9]+)\/?/.exec(value);
			if (sMatches && sMatches[1] && sMatches[2]) {
				out.push(
					new Entity('iframe', {
						src: `https://open.spotify.com/embed/${encodeURIComponent(
							sMatches[1],
						)}/${encodeURIComponent(sMatches[2])}`,
					}),
				);
			}
		}

		{
			const tMatches = /https?:\/\/(?:www\.)?twitch\.tv\/([a-z0-9]+)/.exec(value);
			if (tMatches && tMatches[1]) {
				out.push(
					new Entity('iframe', {
						src: `https://player.twitch.tv/?channel=${encodeURIComponent(
							tMatches[1],
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
