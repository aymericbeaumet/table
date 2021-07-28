<script>
	import gauntletPNG from '$lib/assets/gauntlet.png';
	import gauntletGlowPNG from '$lib/assets/gauntlet-glow.png';

	export let fps;

	export let x;
	export let y;
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
	  transition-duration: ${1 / fps}s;
	  transition-timing-function: linear;
	  left: ${x}px;
	  top: ${y}px;
	  pointer-events: none;
	  user-select: none;
	`}
/>
