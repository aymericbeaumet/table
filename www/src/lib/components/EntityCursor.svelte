<script>
	export let fps;

	export let x;
	export let y;
	export let isDown = false;
</script>

<img
	src={isDown ? '/assets/gauntlet-glow.png' : '/assets/gauntlet.png'}
	alt="cursor"
	style={`
    z-index: 1000;
	  width: 32px;
	  height: 32px;
	  position: absolute;
	  transition-property: top, left;
	  transition-duration: ${1 / fps}s;
	  transition-timing-function: linear;
	  left: ${x}px;
	  top: ${y}px;
	`}
/>
