export enum AmbianceID {
	COFFEE = 'coffee',
	FANTASY = 'fantasy',
}

class Ambiance {
	constructor(
		public name: string,
		public backgroundURL: string,
		public soundURL: string,
		public posterURL: string,
		public cursorURL: string,
		public clickURL: string,
		public loginURL: string,
		public logoutURL: string,
	) {}
}

const ambiances: { [key in AmbianceID]: Ambiance } = {
	[AmbianceID.COFFEE]: coffee(),
	[AmbianceID.FANTASY]: fantasy(),
};
export default ambiances;

/* Coffee */

import coffeeBackground from './coffee/background.jpeg';
import coffeeSound from './coffee/sound.mp3';
import coffeePoster from './coffee/poster.png';
import coffeeCursor from './coffee/cursor.png';
import coffeeClick from './coffee/click.png';
import coffeeLogin from './coffee/login.mp3';
import coffeeLogout from './coffee/logout.mp3';

function coffee() {
	return new Ambiance(
		'Coffee',
		coffeeBackground,
		coffeeSound,
		coffeePoster,
		coffeeCursor,
		coffeeClick,
		coffeeLogin,
		coffeeLogout,
	);
}

/* Fantasy */

import fantasyBackground from './fantasy/background.jpeg';
import fantasySound from './fantasy/sound.mp3';
import fantasyPoster from './fantasy/poster.png';
import fantasyCursor from './fantasy/cursor.png';
import fantasyClick from './fantasy/click.png';
import fantasyLogin from './fantasy/login.mp3';
import fantasyLogout from './fantasy/logout.mp3';

function fantasy() {
	return new Ambiance(
		'Fantasy',
		fantasyBackground,
		fantasySound,
		fantasyPoster,
		fantasyCursor,
		fantasyClick,
		fantasyLogin,
		fantasyLogout,
	);
}
