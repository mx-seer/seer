<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<div class="min-h-screen bg-base-200">
	<div class="navbar bg-base-100 shadow-lg">
		<div class="flex-1">
			<a href="/" class="btn btn-ghost text-xl">
				<span class="text-primary">Seer</span>
			</a>
		</div>
		<div class="flex-none">
			<ul class="menu menu-horizontal px-1">
				<li><a href="/">Dashboard</a></li>
				<li><a href="/sources">Sources</a></li>
				<li><a href="/reports">Reports</a></li>
			</ul>
		</div>
	</div>

	<main class="container mx-auto p-4">
		{@render children()}
	</main>

	<footer class="footer footer-center p-4 bg-base-100 text-base-content">
		<aside>
			<p>"The Seer sees what others miss."</p>
		</aside>
	</footer>
</div>
