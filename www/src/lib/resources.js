import backgroundImageFantasy from '$lib/assets/background-images/fantasy.png';
import backgroundImageGrass from '$lib/assets/background-images/grass.jpeg';
import backgroundImagePebble from '$lib/assets/background-images/pebble.jpeg';
import backgroundImageSand from '$lib/assets/background-images/sand.jpeg';
import backgroundImageSnow from '$lib/assets/background-images/snow.jpeg';
import backgroundImageWater from '$lib/assets/background-images/water.jpeg';
import backgroundImageWooden from '$lib/assets/background-images/wooden.jpeg';

export const backgroundImages = [
	{ id: 'fantasy', name: 'Fantasy', url: backgroundImageFantasy },
	{ id: 'grass', name: 'Grass', url: backgroundImageGrass },
	{ id: 'pebble', name: 'Pebble', url: backgroundImagePebble },
	{ id: 'sand', name: 'Sand', url: backgroundImageSand },
	{ id: 'snow', name: 'Snow', url: backgroundImageSnow },
	{ id: 'water', name: 'Water', url: backgroundImageWater },
	{ id: 'wooden', name: 'Wooden', url: backgroundImageWooden },
];

import backgroundSoundCampfire from '$lib/assets/background-sounds/campfire.mp3';
import backgroundSoundMedieval from '$lib/assets/background-sounds/medieval.mp3';
import backgroundSoundPub from '$lib/assets/background-sounds/pub.mp3';
import backgroundSoundRain from '$lib/assets/background-sounds/rain.mp3';

export const backgroundSounds = [
	{ id: 'silence', name: 'Silence', url: '' },
	{ id: 'campfire', name: 'Campfire', url: backgroundSoundCampfire },
	{ id: 'medieval', name: 'Medieval', url: backgroundSoundMedieval },
	{ id: 'pub', name: 'Pub', url: backgroundSoundPub },
	{ id: 'rain', name: 'Rain', url: backgroundSoundRain },
];
