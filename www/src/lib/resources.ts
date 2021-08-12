interface ResourcesByID {
	[key: string]: Resource;
}

interface Resource {
	name: string;
	url: string;
}

import backgroundImageDesk from '$lib/assets/background-images/desk.jpeg';
import backgroundImageFantasy from '$lib/assets/background-images/fantasy.jpeg';
import backgroundImageLeaves from '$lib/assets/background-images/leaves.jpeg';
import backgroundImagePlants from '$lib/assets/background-images/plants.jpeg';
import backgroundImageTea from '$lib/assets/background-images/tea.jpeg';
import backgroundImageWood from '$lib/assets/background-images/wood.jpeg';

export const backgroundImages: ResourcesByID = {
	desk: { name: 'Desk', url: backgroundImageDesk },
	tea: { name: 'Tea', url: backgroundImageTea },
	leaves: { name: 'Leaves', url: backgroundImageLeaves },
	plants: { name: 'Plants', url: backgroundImagePlants },
	wood: { name: 'Wood', url: backgroundImageWood },
	fantasy: { name: 'Fantasy', url: backgroundImageFantasy },
};

import backgroundSoundCampfire from '$lib/assets/background-sounds/campfire.mp3';
import backgroundSoundMedieval from '$lib/assets/background-sounds/medieval.mp3';
import backgroundSoundPub from '$lib/assets/background-sounds/pub.mp3';
import backgroundSoundRain from '$lib/assets/background-sounds/rain.mp3';

export const backgroundSounds: ResourcesByID = {
	silence: { name: 'Silence', url: '' },
	campfire: { name: 'Campfire', url: backgroundSoundCampfire },
	rain: { name: 'Rain', url: backgroundSoundRain },
	pub: { name: 'Pub', url: backgroundSoundPub },
	medieval: { name: 'Medieval', url: backgroundSoundMedieval },
};
