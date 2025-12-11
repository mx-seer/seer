<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getPrompts,
		getPromptContent,
		type Prompt
	} from '$lib/api';

	let prompts: Prompt[] = $state([]);
	let loading = $state(true);
	let selectedPrompt = $state('');
	let copySuccess = $state(false);
	let promptModal: HTMLDialogElement | null = $state(null);
	let currentPromptCount = $state(0);

	onMount(async () => {
		await loadPrompts();
	});

	async function loadPrompts() {
		loading = true;
		try {
			prompts = await getPrompts();
		} catch (e) {
			console.error('Failed to load prompts:', e);
		}
		loading = false;
	}

	async function handleCopyPrompt(prompt: Prompt) {
		try {
			const content = await getPromptContent(prompt.id);
			selectedPrompt = content;
			currentPromptCount = prompt.opportunity_count;
			promptModal?.showModal();
		} catch (e) {
			console.error('Failed to get prompt:', e);
		}
	}

	async function copyToClipboard() {
		try {
			await navigator.clipboard.writeText(selectedPrompt);
			copySuccess = true;
			setTimeout(() => copySuccess = false, 2000);
		} catch (e) {
			console.error('Failed to copy to clipboard:', e);
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getRelativeTime(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);

		if (diffMins < 60) return `${diffMins} minutes ago`;
		if (diffHours < 24) return `${diffHours} hours ago`;
		if (diffDays === 1) return 'Yesterday';
		return `${diffDays} days ago`;
	}
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-white mb-1">Prompts</h1>
			<p class="text-zinc-400 text-sm">Generated AI prompts for market opportunity analysis</p>
		</div>
	</div>

	<!-- Info Card -->
	<div class="info-banner">
		<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
		</svg>
		<span class="text-zinc-300">
			Prompts are generated from the Dashboard when you select opportunities. Copy them for use with ChatGPT, Claude, or other AI assistants.
		</span>
	</div>

	<!-- Loading State -->
	{#if loading}
		<div class="flex justify-center items-center p-12">
			<div class="loading-spinner"></div>
		</div>
	{:else if prompts.length === 0}
		<!-- Empty State -->
		<div class="empty-state">
			<div class="empty-icon">
				<svg class="w-8 h-8 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
			<h2 class="text-xl font-semibold text-white mb-2">No Prompts Yet</h2>
			<p class="text-zinc-400 mb-6">Select opportunities from the Dashboard and click "Generate AI Prompt" to create your first analysis prompt.</p>
			<a href="/" class="btn btn-primary btn-sm">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<rect x="3" y="3" width="7" height="7" rx="1" stroke-width="2"/>
					<rect x="14" y="3" width="7" height="7" rx="1" stroke-width="2"/>
					<rect x="3" y="14" width="7" height="7" rx="1" stroke-width="2"/>
					<rect x="14" y="14" width="7" height="7" rx="1" stroke-width="2"/>
				</svg>
				Go to Dashboard
			</a>
		</div>
	{:else}
		<!-- Prompts Table -->
		<div class="prompts-table">
			<!-- Table Header -->
			<div class="table-header">
				<div class="col-span-4">Prompt</div>
				<div class="col-span-2">Opportunities</div>
				<div class="col-span-3">Generated</div>
				<div class="col-span-3"></div>
			</div>

			<!-- Table Rows -->
			<div class="divide-y divide-zinc-800">
				{#each prompts as prompt}
					<div class="table-row">
						<div class="col-span-4">
							<div class="flex items-center gap-3">
								<div class="prompt-icon">
									<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
									</svg>
								</div>
								<div>
									<h3 class="text-white font-medium text-sm">Prompt #{prompt.id}</h3>
									<p class="text-zinc-500 text-xs">{getRelativeTime(prompt.created_at)}</p>
								</div>
							</div>
						</div>
						<div class="col-span-2">
							<span class="text-white font-medium">{prompt.opportunity_count}</span>
							<span class="text-zinc-500 text-sm"> found</span>
						</div>
						<div class="col-span-3">
							<span class="text-zinc-400 text-sm">{formatDate(prompt.created_at)}</span>
						</div>
						<div class="col-span-3 flex items-center justify-end gap-2">
							<button
								class="btn btn-outline btn-primary btn-sm"
								onclick={() => handleCopyPrompt(prompt)}
								title="Copy as AI Prompt"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
								</svg>
								Copy Prompt
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Prompt Modal (daisyUI) -->
<dialog bind:this={promptModal} class="modal">
	<div class="modal-box bg-seer-surface border border-seer-border max-w-4xl p-0">
		<!-- Modal Header -->
		<div class="flex items-center justify-between p-6 border-b border-seer-border">
			<div class="flex items-center gap-3">
				<div class="p-2 rounded-lg bg-purple-500/10">
					<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path d="M13 10V3L4 14h7v7l9-11h-7z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
					</svg>
				</div>
				<div>
					<h2 class="text-lg font-semibold text-white">AI Analysis Prompt</h2>
					<p class="text-zinc-500 text-sm">{currentPromptCount} opportunities included</p>
				</div>
			</div>
			<form method="dialog">
				<button class="text-zinc-400 hover:text-white transition-colors p-1" aria-label="Close modal">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path d="M6 18L18 6M6 6l12 12" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
					</svg>
				</button>
			</form>
		</div>

		<!-- Modal Body -->
		<div class="p-6">
			<div class="bg-seer-elevated rounded-lg p-4 max-h-[50vh] overflow-y-auto">
				<pre class="text-zinc-300 text-sm whitespace-pre-wrap font-mono">{selectedPrompt}</pre>
			</div>
		</div>

		<!-- Modal Footer -->
		<div class="flex items-center justify-between p-6 border-t border-seer-border bg-seer-elevated/30">
			<p class="text-zinc-500 text-sm">Copy this prompt and paste it into Claude, ChatGPT, or any AI assistant</p>
			<div class="flex items-center gap-3">
				<form method="dialog">
					<button class="btn btn-ghost btn-sm">Close</button>
				</form>
				<button
					type="button"
					class="btn btn-primary btn-sm"
					onclick={copyToClipboard}
				>
					{#if copySuccess}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path d="M5 13l4 4L19 7" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
						Copied!
					{:else}
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<rect x="9" y="9" width="13" height="13" rx="2" stroke-width="2"/>
							<path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" stroke-width="2"/>
						</svg>
						Copy Prompt
					{/if}
				</button>
			</div>
		</div>
	</div>
	<form method="dialog" class="modal-backdrop">
		<button aria-label="Close modal">close</button>
	</form>
</dialog>

<style>
	.info-banner {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem 1.25rem;
		background: linear-gradient(135deg, rgba(139, 92, 246, 0.1), rgba(124, 58, 237, 0.05));
		border: 1px solid rgba(139, 92, 246, 0.2);
		border-radius: 0.75rem;
	}

	.prompts-table {
		background-color: var(--seer-surface);
		border: 1px solid var(--seer-border);
		border-radius: 0.75rem;
		overflow: hidden;
	}

	.table-header {
		display: grid;
		grid-template-columns: repeat(12, minmax(0, 1fr));
		gap: 1rem;
		padding: 0.75rem 1.5rem;
		background-color: rgba(39, 39, 42, 0.5);
		border-bottom: 1px solid var(--seer-border);
		color: #A1A1AA;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.table-row {
		display: grid;
		grid-template-columns: repeat(12, minmax(0, 1fr));
		gap: 1rem;
		padding: 1rem 1.5rem;
		align-items: center;
		transition: background-color 0.2s ease;
	}

	.table-row:hover {
		background-color: #1F1F23;
	}

	.prompt-icon {
		width: 2.5rem;
		height: 2.5rem;
		background-color: rgba(139, 92, 246, 0.1);
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.empty-state {
		background-color: var(--seer-surface);
		border: 1px solid var(--seer-border);
		border-radius: 0.75rem;
		padding: 3rem;
		text-align: center;
	}

	.empty-icon {
		width: 4rem;
		height: 4rem;
		margin: 0 auto 1rem;
		background-color: var(--seer-elevated);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
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
