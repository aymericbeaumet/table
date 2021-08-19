import Entity from '$lib/entity';
import iframeURL from '$lib/assets/entities/iframe.png';
import soundcloudURL from '$lib/assets/entities/soundcloud.png';
import spotifyURL from '$lib/assets/entities/spotify.png';
import textURL from '$lib/assets/entities/text.png';
import twitchURL from '$lib/assets/entities/twitch.png';
import vimeoURL from '$lib/assets/entities/vimeo.png';
import youtubeURL from '$lib/assets/entities/youtube.png';

export function getIconURL(entity: Entity): string {
	switch (entity.kind) {
		case 'iframe':
			switch (entity?.props?.icon) {
				case 'soundcloud':
					return soundcloudURL;
				case 'spotify':
					return spotifyURL;
				case 'twitch':
					return twitchURL;
				case 'vimeo':
					return vimeoURL;
				case 'youtube':
					return youtubeURL;
				default:
					return iframeURL;
			}

		case 'text':
			return textURL;

		default:
			throw new Error(`implementation error: unexpected kind "${entity.kind}"`);
	}
}
