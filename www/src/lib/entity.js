import getVideoId from 'get-video-id';
import EntityText from './components/EntityText.svelte';
import EntityIframe from './components/EntityIframe.svelte';

export function inferEntity(value) {
	const { service, id } = getVideoId(value);
	switch (service) {
		case 'vimeo':
			return [EntityIframe, { src: `https://player.vimeo.com/video/${id}` }];
		case 'youtube':
			return [EntityIframe, { src: `https://www.youtube-nocookie.com/embed/${id}` }];
	}

	const matches = /https?:\/\/(?:www\.)?twitch\.tv\/([a-z0-9]+)/.exec(value);
	if (matches && matches.length === 2) {
		return [
			EntityIframe,
			{
				src: `https://player.twitch.tv/?channel=${matches[1]}&parent=${
					import.meta.env.VITE_HOSTNAME
				}`,
			},
		];
	}

	return [EntityText, { value }];
}
