<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';

	let { children } = $props();
</script>

<div class="min-h-screen flex flex-col bg-seer-bg">
	<!-- Top Accent Line -->
	<div class="top-accent"></div>

	<!-- Navbar -->
	<nav class="bg-seer-surface border-b border-seer-border">
		<div class="max-w-7xl mx-auto px-6">
			<div class="flex items-center justify-between h-16">
				<!-- Logo + Tagline -->
				<div class="flex items-center gap-3">
					<a href="/" class="flex items-center gap-2">
						<!-- Crystal Ball Logo SVG with glow filter like seer-ui -->
						<svg width="36" height="36" viewBox="-4 -4 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
							<defs>
								<linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
									<stop offset="0%" style="stop-color:#A78BFA"/>
									<stop offset="100%" style="stop-color:#7C3AED"/>
								</linearGradient>
								<filter id="glowNav" x="-50%" y="-50%" width="200%" height="200%">
									<feGaussianBlur stdDeviation="2" result="coloredBlur"/>
									<feMerge>
										<feMergeNode in="coloredBlur"/>
										<feMergeNode in="SourceGraphic"/>
									</feMerge>
								</filter>
							</defs>
							<g filter="url(#glowNav)">
								<rect x="1" y="1" width="30" height="30" rx="8" fill="none" stroke="url(#logoGradient)" stroke-width="1.5"/>
								<circle cx="16" cy="13" r="7" stroke="url(#logoGradient)" stroke-width="1.5" fill="none"/>
								<path d="M12 10C13 8.5 14.5 8 16 8" stroke="white" stroke-width="1" stroke-linecap="round" opacity="0.5"/>
								<path d="M11 22C11 22 13 24 16 24C19 24 21 22 21 22" stroke="url(#logoGradient)" stroke-width="1.5" stroke-linecap="round"/>
								<path d="M12.5 22V20" stroke="url(#logoGradient)" stroke-width="1.5" stroke-linecap="round"/>
								<path d="M19.5 22V20" stroke="url(#logoGradient)" stroke-width="1.5" stroke-linecap="round"/>
							</g>
						</svg>
						<span class="font-display text-xl font-semibold text-white">Seer</span>
					</a>
					<span class="text-zinc-500 text-sm hidden sm:block">See what others miss</span>
				</div>

				<!-- Navigation -->
				<div class="flex items-center gap-6">
					<a
						href="/"
						class="nav-link"
						class:nav-active={$page.url.pathname === '/'}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<rect x="3" y="3" width="7" height="7" rx="1" stroke-width="2"/>
							<rect x="14" y="3" width="7" height="7" rx="1" stroke-width="2"/>
							<rect x="3" y="14" width="7" height="7" rx="1" stroke-width="2"/>
							<rect x="14" y="14" width="7" height="7" rx="1" stroke-width="2"/>
						</svg>
						Dashboard
					</a>
					<a
						href="/reports"
						class="nav-link"
						class:nav-active={$page.url.pathname === '/reports'}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path d="M13 10V3L4 14h7v7l9-11h-7z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
						Prompts
					</a>
					<a
						href="/sources"
						class="nav-link"
						class:nav-active={$page.url.pathname === '/sources'}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<circle cx="12" cy="12" r="2" stroke-width="2"/>
							<path d="M16.24 7.76a6 6 0 010 8.49m-8.48-.01a6 6 0 010-8.49m11.31-2.82a10 10 0 010 14.14m-14.14 0a10 10 0 010-14.14" stroke-width="2" stroke-linecap="round"/>
						</svg>
						Sources
					</a>
				</div>
			</div>
		</div>
	</nav>

	<!-- Main Content -->
	<main class="flex-1 max-w-7xl w-full mx-auto px-6 py-8">
		{@render children()}
	</main>

	<!-- Footer -->
	<footer class="border-t border-seer-border">
		<div class="max-w-7xl mx-auto px-6 py-6">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<!-- Mini crystal ball logo with glow filter like seer-ui -->
					<svg width="24" height="24" viewBox="-4 -4 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
						<defs>
							<linearGradient id="footerGradient" x1="0%" y1="0%" x2="100%" y2="100%">
								<stop offset="0%" style="stop-color:#A78BFA"/>
								<stop offset="100%" style="stop-color:#7C3AED"/>
							</linearGradient>
							<filter id="glowFooter" x="-50%" y="-50%" width="200%" height="200%">
								<feGaussianBlur stdDeviation="2" result="coloredBlur"/>
								<feMerge>
									<feMergeNode in="coloredBlur"/>
									<feMergeNode in="SourceGraphic"/>
								</feMerge>
							</filter>
						</defs>
						<g filter="url(#glowFooter)">
							<rect x="1" y="1" width="30" height="30" rx="8" fill="none" stroke="url(#footerGradient)" stroke-width="1.5"/>
							<circle cx="16" cy="13" r="7" stroke="url(#footerGradient)" stroke-width="1.5" fill="none"/>
							<path d="M12 10C13 8.5 14.5 8 16 8" stroke="white" stroke-width="1" stroke-linecap="round" opacity="0.5"/>
							<path d="M11 22C11 22 13 24 16 24C19 24 21 22 21 22" stroke="url(#footerGradient)" stroke-width="1.5" stroke-linecap="round"/>
							<path d="M12.5 22V20" stroke="url(#footerGradient)" stroke-width="1.5" stroke-linecap="round"/>
							<path d="M19.5 22V20" stroke="url(#footerGradient)" stroke-width="1.5" stroke-linecap="round"/>
						</g>
					</svg>
					<span class="font-display text-sm font-semibold text-zinc-300">Seer</span>
					<span class="text-zinc-500 text-sm">by</span>
					<a href="https://mendex.io" class="text-purple-400 hover:text-purple-300 transition-colors text-sm">Mendex</a>
				</div>
				<div class="text-zinc-600 text-sm">
					v0.1.0
				</div>
			</div>
		</div>
	</footer>
</div>
