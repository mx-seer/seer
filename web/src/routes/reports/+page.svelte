<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getReports,
		generateReport,
		getReportPrompt,
		type Report
	} from '$lib/api';

	let reports: Report[] = $state([]);
	let loading = $state(true);
	let generating = $state(false);
	let selectedPrompt = $state('');
	let showPromptModal = $state(false);

	onMount(async () => {
		await loadReports();
	});

	async function loadReports() {
		loading = true;
		try {
			reports = await getReports();
		} catch (e) {
			console.error('Failed to load reports:', e);
		}
		loading = false;
	}

	async function handleGenerate() {
		generating = true;
		try {
			const report = await generateReport();
			reports = [report, ...reports];
		} catch (e) {
			console.error('Failed to generate report:', e);
		}
		generating = false;
	}

	async function handleCopyPrompt(id: number) {
		try {
			const prompt = await getReportPrompt(id);
			selectedPrompt = prompt;
			showPromptModal = true;
		} catch (e) {
			console.error('Failed to get prompt:', e);
		}
	}

	async function copyToClipboard() {
		try {
			await navigator.clipboard.writeText(selectedPrompt);
			showPromptModal = false;
		} catch (e) {
			console.error('Failed to copy to clipboard:', e);
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString();
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-2xl font-bold">Reports</h1>
		<button
			class="btn btn-primary"
			onclick={handleGenerate}
			disabled={generating}
		>
			{#if generating}
				<span class="loading loading-spinner loading-sm"></span>
				Generating...
			{:else}
				Generate Report
			{/if}
		</button>
	</div>

	<div class="alert alert-info">
		<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
		</svg>
		<span>Reports analyze recent opportunities and can be copied as prompts for external AI tools.</span>
	</div>

	{#if loading}
		<div class="flex justify-center p-8">
			<span class="loading loading-spinner loading-lg"></span>
		</div>
	{:else if reports.length === 0}
		<div class="card bg-base-100 shadow-xl">
			<div class="card-body items-center text-center">
				<h2 class="card-title">No Reports Yet</h2>
				<p>Generate your first report to get AI-ready analysis of market opportunities.</p>
				<div class="card-actions">
					<button class="btn btn-primary" onclick={handleGenerate} disabled={generating}>
						Generate First Report
					</button>
				</div>
			</div>
		</div>
	{:else}
		<div class="space-y-4">
			{#each reports as report}
				<div class="card bg-base-100 shadow-xl">
					<div class="card-body">
						<div class="flex justify-between items-start">
							<div>
								<h2 class="card-title">
									Report #{report.id}
									<div class="badge badge-secondary">{report.opportunity_count} opportunities</div>
								</h2>
								<p class="text-sm text-base-content/60">
									Generated: {formatDate(report.generated_at)}
								</p>
								{#if report.period_start && report.period_end}
									<p class="text-sm text-base-content/60">
										Period: {formatDate(report.period_start)} - {formatDate(report.period_end)}
									</p>
								{/if}
							</div>
							<button
								class="btn btn-outline btn-sm"
								onclick={() => handleCopyPrompt(report.id)}
							>
								Copy as Prompt
							</button>
						</div>

						{#if report.summary}
							<div class="divider"></div>
							<div class="prose max-w-none">
								<pre class="whitespace-pre-wrap text-sm bg-base-200 p-4 rounded-lg">{report.summary}</pre>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Prompt Modal -->
{#if showPromptModal}
	<div class="modal modal-open">
		<div class="modal-box max-w-4xl">
			<h3 class="font-bold text-lg">Copy Prompt for AI</h3>
			<p class="py-2 text-sm text-base-content/60">
				Copy this prompt to use with ChatGPT, Claude, or other AI assistants.
			</p>

			<div class="form-control mt-4">
				<textarea
					class="textarea textarea-bordered h-96 font-mono text-sm"
					readonly
					value={selectedPrompt}
				></textarea>
			</div>

			<div class="modal-action">
				<button class="btn" onclick={() => (showPromptModal = false)}>Cancel</button>
				<button class="btn btn-primary" onclick={copyToClipboard}>
					Copy to Clipboard
				</button>
			</div>
		</div>
		<button class="modal-backdrop" onclick={() => (showPromptModal = false)} aria-label="Close modal"></button>
	</div>
{/if}
