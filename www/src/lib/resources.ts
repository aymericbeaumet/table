interface ResourcesByID {
	[key: string]: Resource;
}

interface Resource {
	name: string;
	url: string;
}

import backgroundImageFantasy from '$lib/assets/background-images/fantasy.png';
import backgroundImageGrass from '$lib/assets/background-images/grass.jpeg';
import backgroundImagePebble from '$lib/assets/background-images/pebble.jpeg';
import backgroundImageSand from '$lib/assets/background-images/sand.jpeg';
import backgroundImageSnow from '$lib/assets/background-images/snow.jpeg';
import backgroundImageWater from '$lib/assets/background-images/water.jpeg';
import backgroundImageWooden from '$lib/assets/background-images/wooden.jpeg';

export const backgroundImages: ResourcesByID = {
	fantasy: { name: 'Fantasy', url: backgroundImageFantasy },
	grass: { name: 'Grass', url: backgroundImageGrass },
	pebble: { name: 'Pebble', url: backgroundImagePebble },
	sand: { name: 'Sand', url: backgroundImageSand },
	snow: { name: 'Snow', url: backgroundImageSnow },
	water: { name: 'Water', url: backgroundImageWater },
	wooden: { name: 'Wooden', url: backgroundImageWooden },
};

import backgroundSoundCampfire from '$lib/assets/background-sounds/campfire.mp3';
import backgroundSoundMedieval from '$lib/assets/background-sounds/medieval.mp3';
import backgroundSoundPub from '$lib/assets/background-sounds/pub.mp3';
import backgroundSoundRain from '$lib/assets/background-sounds/rain.mp3';

export const backgroundSounds: ResourcesByID = {
	silence: { name: 'Silence', url: '' },
	campfire: { name: 'Campfire', url: backgroundSoundCampfire },
	medieval: { name: 'Medieval', url: backgroundSoundMedieval },
	pub: { name: 'Pub', url: backgroundSoundPub },
	rain: { name: 'Rain', url: backgroundSoundRain },
};
