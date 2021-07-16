<script>
	export let name;

	function copyJoiningInfo() {
		navigator.clipboard.writeText(window.location.toString());
	}
</script>

<header>
	<div class="title" on:click={copyJoiningInfo} title="Copy joining info">
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

	.title {
		cursor: pointer;
		font-size: 15px;
		font-weight: normal;
	}

	h1 {
		display: inline;
	}
</style>
