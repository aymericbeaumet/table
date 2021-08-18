import Entity from '$lib/entity';
import iframeURL from '$lib/assets/entities/iframe.png';
import textURL from '$lib/assets/entities/text.png';

export function getIconURL(entity: Entity): string {
	switch (entity.kind) {
		case 'iframe':
			return iframeURL;
		case 'text':
			return textURL;
		default:
			throw new Error(`implementation error: unexpected kind "${entity.kind}"`);
	}
}
