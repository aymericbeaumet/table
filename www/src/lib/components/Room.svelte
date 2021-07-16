<script>
	export let name;
</script>

<header>
	<div>
		<h1>{name}</h1>
		<span>[copy]</span>
	</div>

	<div>
		<span>[settings]</span>
	</div>
</header>

<main />

<style>
	header {
		border-bottom: 1px solid black;
		display: flex;
		justify-content: space-between;
		padding: 10px;
	}

	h1 {
		display: inline;
	}
</style>
