import getVideoId from 'get-video-id';
import EntityEphemeralText from '$lib/components/EntityEphemeralText.svelte';
import EntityPersistentText from '$lib/components/EntityPersistentText.svelte';
import EntityIframe from '$lib/components/EntityIframe.svelte';

export function inferEntities(value) {
	const out = [];

	if (value.length === 0) {
		return out;
	}

	const { service, id } = getVideoId(value);
	switch (service) {
		case 'vimeo':
			out.push([
				'vimeo',
				EntityIframe,
				{ src: `https://player.vimeo.com/video/${encodeURIComponent(id)}` },
			]);
		case 'youtube':
			out.push([
				'youtube',
				EntityIframe,
				{ src: `https://www.youtube-nocookie.com/embed/${encodeURIComponent(id)}` },
			]);
	}

	const matches = /https?:\/\/(?:www\.)?twitch\.tv\/([a-z0-9]+)/.exec(value);
	if (matches && matches.length === 2) {
		out.push([
			'twitch',
			EntityIframe,
			{
				src: `https://player.twitch.tv/?channel=${encodeURIComponent(
					matches[1],
				)}&parent=${encodeURIComponent(window.location.hostname)}`,
			},
		]);
	}

	out.push(['persistent', EntityPersistentText, { value }]);
	out.push(['ephemeral', EntityEphemeralText, { value }]);

	return out;
}
