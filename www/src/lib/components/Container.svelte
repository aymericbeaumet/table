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
				<li class="shiny"><a href="/start">Start a table</a></li>
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
		min-height: 100vh;
		display: flex;
		flex-direction: column;
	}

	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		border-bottom: 1px solid black;
		padding: 15px 20px;
	}

	header ul {
		display: flex;
	}

	header a {
		padding: 5px;
	}

	header li {
		margin: 0 2px;
	}

	header li.shiny a {
		background: orange;
		border: 1px solid orange;
		color: white;
		border-radius: 5px;
	}

	header li.shiny a:hover {
		border: 1px solid black;
		color: black;
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
		flex: 1;
		padding: 20px;
	}

	footer {
		padding: 30px;
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
