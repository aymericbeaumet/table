<script>
	import gauntletPNG from '$lib/assets/gauntlet.png';
	import gauntletGlowPNG from '$lib/assets/gauntlet-glow.png';
	import { MAX_FPS } from '$lib/env';

	export let clientX: number;
	export let clientY: number;
	export let isDown = false;
</script>

<img
	src={isDown ? gauntletGlowPNG : gauntletPNG}
	alt="cursor"
	style={`
	  z-index: 500;
	  width: 32px;
	  height: 32px;
	  position: absolute;
	  transition-property: top, left;
	  transition-duration: ${1 / MAX_FPS}s;
	  transition-timing-function: linear;
	  left: ${clientX}px;
	  top: ${clientY}px;
	  pointer-events: none;
	  user-select: none;
	`}
/>
