import getVideoId from 'get-video-id';
import EntityEphemeralText from '$lib/components/EntityEphemeralText.svelte';
import EntityPersistentText from '$lib/components/EntityPersistentText.svelte';
import EntityIframe from '$lib/components/EntityIframe.svelte';

export function inferEntities(value) {
	const out = [];

	const { service, id } = getVideoId(value);
	switch (service) {
		case 'vimeo':
			out.push([EntityIframe, { src: `https://player.vimeo.com/video/${id}` }]);
		case 'youtube':
			out.push([EntityIframe, { src: `https://www.youtube-nocookie.com/embed/${id}` }]);
	}

	const matches = /https?:\/\/(?:www\.)?twitch\.tv\/([a-z0-9]+)/.exec(value);
	if (matches && matches.length === 2) {
		out.push([
			EntityIframe,
			{
				src: `https://player.twitch.tv/?channel=${encodeURIComponent(
					matches[1],
				)}&parent=${encodeURIComponent(window.location.hostname)}`,
			},
		]);
	}

	if (value.length > 0) {
		out.push([EntityPersistentText, { value }]);
		out.push([EntityEphemeralText, { value }]);
	}

	return out;
}
