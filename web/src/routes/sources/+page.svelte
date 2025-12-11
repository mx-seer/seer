<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getSources,
		toggleSource,
		type Source
	} from '$lib/api';

	let sources: Source[] = $state([]);
	let loading = $state(true);

	// Source icons by type
	const sourceIcons: Record<string, { bg: string; color: string; label: string }> = {
		hackernews: { bg: 'rgba(249, 115, 22, 0.1)', color: '#FB923C', label: 'Y' },
		github: { bg: 'rgba(161, 161, 170, 0.1)', color: '#A1A1AA', label: '' },
		npm: { bg: 'rgba(239, 68, 68, 0.1)', color: '#F87171', label: 'npm' },
		devto: { bg: 'rgba(161, 161, 170, 0.1)', color: '#A1A1AA', label: 'DEV' },
		rss: { bg: 'rgba(139, 92, 246, 0.1)', color: '#A78BFA', label: '' },
		reddit: { bg: 'rgba(249, 115, 22, 0.1)', color: '#FB923C', label: '' },
		producthunt: { bg: 'rgba(249, 115, 22, 0.1)', color: '#FB923C', label: '' }
	};

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		try {
			sources = await getSources();
		} catch (e) {
			console.error('Failed to load sources:', e);
		}
		loading = false;
	}

	async function handleToggle(id: number) {
		try {
			const updated = await toggleSource(id);
			sources = sources.map((s) => (s.id === id ? updated : s));
		} catch (e) {
			console.error('Failed to toggle source:', e);
		}
	}

	function getActiveCount(): number {
		return sources.filter((s) => s.enabled).length;
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-white mb-1">Sources</h1>
			<p class="text-zinc-400 text-sm">Configure where Seer looks for opportunities</p>
		</div>
		<div class="flex items-center gap-4">
			{#if sources.length > 0}
				<div class="flex items-center gap-2 text-sm">
					<span class="text-zinc-500">{getActiveCount()} of {sources.length} active</span>
					{#if getActiveCount() > 0}
						<span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
					{:else}
						<span class="w-2 h-2 bg-red-400 rounded-full"></span>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Info Banner -->
	<div class="info-banner">
		<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
		</svg>
		<span class="text-zinc-300">
			Sources are automatically scanned for new opportunities. Toggle them on or off to customize your feed.
		</span>
	</div>

	<!-- Loading State -->
	{#if loading}
		<div class="flex justify-center items-center p-12">
			<div class="loading-spinner"></div>
		</div>
	{:else}
		<!-- Sources Grid -->
		<div class="grid gap-4 md:grid-cols-2">
			{#each sources as source}
				<div class="source-card" class:source-disabled={!source.enabled}>
					<div class="flex items-start justify-between mb-4">
						<div class="flex items-center gap-3">
							<!-- Source Icon -->
							<div
								class="w-10 h-10 rounded-lg flex items-center justify-center"
								style="background-color: {sourceIcons[source.type]?.bg || 'rgba(139, 92, 246, 0.1)'};"
							>
								{#if source.type === 'github'}
									<svg class="w-5 h-5" style="color: {sourceIcons[source.type]?.color};" fill="currentColor" viewBox="0 0 24 24">
										<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
									</svg>
								{:else if source.type === 'rss'}
									<svg class="w-5 h-5" style="color: {sourceIcons[source.type]?.color};" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 5c7.18 0 13 5.82 13 13M6 11a7 7 0 017 7m-6 0a1 1 0 11-2 0 1 1 0 012 0z" />
									</svg>
								{:else}
									<span class="font-bold text-sm" style="color: {sourceIcons[source.type]?.color};">
										{sourceIcons[source.type]?.label || source.type.charAt(0).toUpperCase()}
									</span>
								{/if}
							</div>
							<div>
								<h3 class="text-white font-semibold">{source.name}</h3>
								<p class="text-zinc-500 text-sm">
									{#if source.url}
										{source.url}
									{:else}
										{source.type}
									{/if}
								</p>
							</div>
						</div>
						<!-- Toggle Switch (daisyUI) -->
						<input
							type="checkbox"
							class="toggle toggle-primary toggle-sm"
							checked={source.enabled}
							onchange={() => handleToggle(source.id)}
						/>
					</div>

					<div class="space-y-3">
						<div class="flex items-center justify-between text-sm">
							<span class="text-zinc-400">Status</span>
							{#if source.enabled}
								<span class="text-green-400 flex items-center gap-1">
									<span class="w-1.5 h-1.5 bg-green-400 rounded-full"></span>
									Active
								</span>
							{:else}
								<span class="text-zinc-500 flex items-center gap-1">
									<span class="w-1.5 h-1.5 bg-zinc-500 rounded-full"></span>
									Disabled
								</span>
							{/if}
						</div>
						{#if source.is_builtin}
							<div class="flex items-center justify-between text-sm">
								<span class="text-zinc-400">Type</span>
								<span class="badge badge-outline badge-primary badge-sm">Built-in</span>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.source-card {
		background-color: var(--seer-surface);
		border: 1px solid var(--seer-border);
		border-radius: 0.75rem;
		padding: 1.5rem;
		transition: all 0.2s ease;
	}

	.source-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 0 30px rgba(139, 92, 246, 0.15);
	}

	.source-disabled {
		opacity: 0.6;
	}

	.source-disabled:hover {
		transform: none;
		box-shadow: none;
	}

	.info-banner {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem 1.25rem;
		background: linear-gradient(135deg, rgba(139, 92, 246, 0.1), rgba(124, 58, 237, 0.05));
		border: 1px solid rgba(139, 92, 246, 0.2);
		border-radius: 0.75rem;
	}

	.loading-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--seer-border);
		border-top-color: var(--purple-500);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}
</style>
