<script>
	import logoURL from '$lib/assets/logo.png';
</script>

<div>
	<header>
		<a class="logo" href="/"><img src={logoURL} alt="logo" />Table</a>

		<nav>
			<ul>
				<li><a href="/blog">Blog</a></li>
				<li><a href="/doc">Documentation</a></li>
			</ul>
		</nav>
	</header>

	<main>
		<slot />
	</main>

	<footer>
		<ul>
			<li><a href="/about">About</a></li>
			<li><a href="/careers">Careers</a></li>
			<li><a href="/terms">Terms</a></li>
			<li><a href="/privacy">Privacy</a></li>
			<li><a href="/security">Security</a></li>
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
		padding: 20px;
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

	header a.logo {
		font-size: 20px;
		text-transform: uppercase;
		text-decoration: none;
		font-weight: bold;
		display: flex;
		align-items: center;
	}

	header a:hover {
		color: orange;
	}

	header img {
		height: 32px;
		width: 32px;
		margin-right: 8px;
	}

	main {
		padding: 20px;
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
