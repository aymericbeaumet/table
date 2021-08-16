<script>
	import Logo from '$lib/components/Logo.svelte';
</script>

<div>
	<header>
		<Logo />

		<nav>
			<ul>
				<li><a href="/blog">Blog</a></li>
				<li><a href="/docs">Docs</a></li>
				<li><a href="/docs/getting-started">Getting Started</a></li>
			</ul>
		</nav>
	</header>

	<main>
		<slot />
	</main>

	<footer>
		<ul>
			<li><a href="/careers">Careers</a></li>
			<li><a href="/terms">Terms</a></li>
			<li><a href="/privacy">Privacy</a></li>
		</ul>
	</footer>
</div>

<style>
	div {
		margin: 0 auto;
		max-width: 1000px;
		height: 100vh;
	}

	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		border-bottom: 1px solid black;
		padding: 30px;
	}

	header ul {
		display: flex;
	}

	header li + li {
		margin-left: 10px;
	}

	header a {
		color: black;
		text-decoration: none;
	}

	main {
		padding: 30px;
	}

	footer {
		padding: 20px;
	}

	footer ul {
		display: flex;
		justify-content: center;
	}

	footer li + li {
		margin-left: 10px;
	}

	footer a {
		color: grey;
		text-decoration: none;
	}

	ul {
		list-style: none;
	}
</style>
