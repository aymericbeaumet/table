<script>
	export let x;
	export let y;
	export let fps;
</script>

<img
	src="/assets/cursor.png"
	alt="cursor"
	style={`
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
