<script>
	export let x;
	export let y;

	console.log({ x, y });
</script>

<img
	src="/assets/cursor.png"
	alt="cursor"
	style={`position: absolute; left: ${x}px; top: ${y}px; width: 32px; height: 32px;`}
/>
