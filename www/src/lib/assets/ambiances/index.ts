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

const ambiances: { [key: string]: Ambiance } = {};

export default ambiances;

/* Office */

import officeBackground from './office/background.jpeg';
import officeSound from './office/sound.mp3';
import officePoster from './office/poster.png';
import officeCursor from './office/cursor.png';
import officeClick from './office/click.png';
import officeLogin from './office/login.mp3';
import officeLogout from './office/logout.mp3';

ambiances['office'] = new Ambiance(
	'Office',
	officeBackground,
	officeSound,
	officePoster,
	officeCursor,
	officeClick,
	officeLogin,
	officeLogout,
);

/* Fantasy */

import fantasyBackground from './fantasy/background.jpeg';
import fantasySound from './fantasy/sound.mp3';
import fantasyPoster from './fantasy/poster.png';
import fantasyCursor from './fantasy/cursor.png';
import fantasyClick from './fantasy/click.png';
import fantasyLogin from './fantasy/login.mp3';
import fantasyLogout from './fantasy/logout.mp3';

ambiances['fantasy'] = new Ambiance(
	'Fantasy',
	fantasyBackground,
	fantasySound,
	fantasyPoster,
	fantasyCursor,
	fantasyClick,
	fantasyLogin,
	fantasyLogout,
);
