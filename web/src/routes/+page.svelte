<script lang="ts">
	import { onMount } from 'svelte';
	import { getOpportunities, getStats, getSources, createPrompt, fetchSources, type Opportunity, type Stats, type Source } from '$lib/api';

	let opportunities: Opportunity[] = $state([]);
	let stats: Stats | null = $state(null);
	let sources: Source[] = $state([]);
	let loading = $state(true);
	let selectedSource = $state('');
	let minScore = $state(60);

	// Search
	let searchQuery = $state('');

	// Pagination
	let currentPage = $state(1);
	let pageSize = $state(5);
	let totalItems = $state(0);

	// Sorting
	type SortColumn = 'score' | 'detected_at' | 'source_type' | 'title';
	type SortDirection = 'asc' | 'desc';
	let sortColumn: SortColumn = $state('detected_at');
	let sortDirection: SortDirection = $state('desc');

	// Detail Modal
	let selectedOpportunity: Opportunity | null = $state(null);
	let detailModal: HTMLDialogElement;

	// Prompt Modal
	let promptModal: HTMLDialogElement;
	let generatedPrompt = $state('');
	let copySuccess = $state(false);

	// Analysis selection
	let analysisSet: Set<number> = $state(new Set());

	// Refetch state
	let refetching = $state(false);

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		try {
			const [opps, s, src] = await Promise.all([
				getOpportunities({
					source: selectedSource || undefined,
					min_score: minScore || undefined,
					limit: 500 // Load more for client-side sorting/pagination
				}),
				getStats({
					source: selectedSource || undefined,
					min_score: minScore || undefined
				}),
				getSources()
			]);
			opportunities = opps;
			stats = s;
			sources = src;
			totalItems = opps.length;
			currentPage = 1; // Reset to first page on filter change
		} catch (e) {
			console.error('Failed to load data:', e);
		}
		loading = false;
	}

	// Handle refetch from sources
	async function handleRefetch() {
		if (refetching) return;
		refetching = true;
		try {
			await fetchSources();
			// Wait a bit then reload data
			setTimeout(async () => {
				await loadData();
				refetching = false;
			}, 3000);
		} catch (e) {
			console.error('Failed to refetch:', e);
			refetching = false;
		}
	}

	// Count enabled sources
	function enabledSourcesCount(): number {
		return sources.filter(s => s.enabled).length;
	}

	// Calculate average score from filtered opportunities
	function filteredAverageScore(): number {
		const filtered = filteredOpportunities();
		if (filtered.length === 0) return 0;
		const sum = filtered.reduce((acc, opp) => acc + opp.score, 0);
		return Math.round(sum / filtered.length);
	}

	// Filter by search query
	function filteredOpportunities(): Opportunity[] {
		if (!searchQuery.trim()) return opportunities;
		const query = searchQuery.toLowerCase();
		return opportunities.filter(
			(opp) =>
				opp.title.toLowerCase().includes(query) ||
				opp.description?.toLowerCase().includes(query) ||
				opp.source_type.toLowerCase().includes(query)
		);
	}

	// Sort opportunities
	function sortedOpportunities(): Opportunity[] {
		return [...filteredOpportunities()].sort((a, b) => {
			let comparison = 0;
			switch (sortColumn) {
				case 'score':
					comparison = a.score - b.score;
					break;
				case 'detected_at':
					comparison = new Date(a.detected_at).getTime() - new Date(b.detected_at).getTime();
					break;
				case 'source_type':
					comparison = a.source_type.localeCompare(b.source_type);
					break;
				case 'title':
					comparison = a.title.localeCompare(b.title);
					break;
			}
			return sortDirection === 'asc' ? comparison : -comparison;
		});
	}

	// Paginated opportunities
	function paginatedOpportunities(): Opportunity[] {
		const sorted = sortedOpportunities();
		const start = (currentPage - 1) * pageSize;
		return sorted.slice(start, start + pageSize);
	}

	// Total filtered items
	function filteredTotal(): number {
		return filteredOpportunities().length;
	}

	// Total pages
	function totalPages(): number {
		return Math.ceil(filteredTotal() / pageSize);
	}

	// Handle page size change
	function handlePageSizeChange() {
		currentPage = 1; // Reset to first page when changing page size
	}

	// Handle sort
	function handleSort(column: SortColumn) {
		if (sortColumn === column) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortColumn = column;
			sortDirection = column === 'detected_at' ? 'desc' : 'desc'; // Default desc for most
		}
	}

	// Pagination handlers
	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages()) {
			currentPage = page;
		}
	}

	function getPageNumbers(): number[] {
		const total = totalPages();
		const current = currentPage;
		const pages: number[] = [];

		if (total <= 5) {
			for (let i = 1; i <= total; i++) pages.push(i);
		} else {
			if (current <= 3) {
				pages.push(1, 2, 3, 4, -1, total);
			} else if (current >= total - 2) {
				pages.push(1, -1, total - 3, total - 2, total - 1, total);
			} else {
				pages.push(1, -1, current - 1, current, current + 1, -1, total);
			}
		}
		return pages;
	}

	function getScoreClass(score: number): string {
		if (score >= 90) return 'score-elite';
		if (score >= 60) return 'score-high';
		if (score >= 31) return 'score-mid';
		return 'score-low';
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffDays = Math.floor(diffHours / 24);

		if (diffHours < 1) return 'Just now';
		if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
		if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	// Sort indicator
	function getSortIndicator(column: SortColumn): string {
		if (sortColumn !== column) return '';
		return sortDirection === 'asc' ? ' ↑' : ' ↓';
	}

	// Modal functions
	function openDetail(opp: Opportunity) {
		selectedOpportunity = opp;
		detailModal?.showModal();
	}

	function toggleAnalysis(id: number) {
		const newSet = new Set(analysisSet);
		if (newSet.has(id)) {
			newSet.delete(id);
		} else {
			newSet.add(id);
		}
		analysisSet = newSet;
	}

	function isInAnalysis(id: number): boolean {
		return analysisSet.has(id);
	}

	// Signal display name
	function formatSignal(signal: string): string {
		return signal.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
	}

	// Generate AI Prompt
	async function generatePrompt() {
		const selectedOpps = opportunities.filter(opp => analysisSet.has(opp.id));
		if (selectedOpps.length === 0) return;

		const prompt = `You are an expert market analyst and startup advisor. Analyze the following ${selectedOpps.length} market opportunities detected by Seer and provide:

1. **Overall Assessment**: Rate the quality of these opportunities as a group (1-10)
2. **Top Recommendations**: Which 2-3 opportunities show the most potential and why?
3. **Market Patterns**: What common themes or trends do you see?
4. **Action Items**: Specific next steps to validate or pursue the best opportunities
5. **Risk Analysis**: Potential challenges or red flags to consider

---

## Opportunities to Analyze:

${selectedOpps.map((opp, i) => `### ${i + 1}. ${opp.title}
- **Score**: ${opp.score}/100
- **Source**: ${opp.source_type}
- **Signals**: ${opp.signals.map(s => formatSignal(s)).join(', ')}
- **Description**: ${opp.description || 'N/A'}
- **Detected**: ${formatDate(opp.detected_at)}
- **Link**: ${opp.source_url}
`).join('\n')}

---

Please provide a detailed analysis focusing on actionable insights for an indie hacker or small startup looking to validate and pursue these opportunities.`;

		generatedPrompt = prompt;
		copySuccess = false;
		promptModal?.showModal();

		// Save prompt
		try {
			await createPrompt({
				opportunity_count: selectedOpps.length,
				content_prompt: prompt
			});
		} catch (e) {
			console.error('Failed to save prompt:', e);
		}
	}

	// Copy prompt to clipboard
	async function copyPrompt() {
		try {
			await navigator.clipboard.writeText(generatedPrompt);
			copySuccess = true;
			setTimeout(() => copySuccess = false, 2000);
		} catch (e) {
			console.error('Failed to copy:', e);
		}
	}

	// Get selected count
	function selectedCount(): number {
		return analysisSet.size;
	}

	// Score breakdown estimation based on signals
	function getScoreBreakdown(signals: string[]): { signal: string; contribution: number }[] {
		const signalWeights: Record<string, number> = {
			'problem_mention': 20,
			'solution_seeking': 25,
			'indie_focus': 15,
			'building_in_public': 15,
			'market_validation': 20,
			'pain_point': 20,
			'feature_request': 15,
			'user_frustration': 18,
			'tool_comparison': 12,
			'budget_mention': 10,
			'urgency': 15,
			'growth_potential': 18,
			'underserved_market': 20,
			'integration_need': 12,
			'automation_desire': 15
		};

		return signals.map(signal => ({
			signal,
			contribution: signalWeights[signal] || 10
		}));
	}
</script>

<div class="space-y-8">
	<!-- Stats Cards - EXACT seer-ui Tailwind classes -->
	{#if stats}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<!-- Total Opportunities -->
			<div class="bg-seer-surface border border-seer-border rounded-lg p-6 card-hover glow-purple">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 rounded-lg bg-purple-500/10">
						<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</div>
					<span class="text-zinc-400 text-sm font-medium">Total Opportunities</span>
				</div>
				<div class="flex items-end justify-between">
					<span class="text-4xl font-bold text-white tabular-nums">{stats.total}</span>
					<span class="text-purple-400 text-sm font-medium">+12%</span>
				</div>
			</div>

			<!-- Today -->
			<div class="bg-seer-surface border border-seer-border rounded-lg p-6 card-hover">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 rounded-lg bg-purple-500/10">
						<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<rect x="3" y="4" width="18" height="18" rx="2" stroke-width="2"/>
							<path d="M16 2v4M8 2v4M3 10h18" stroke-width="2" stroke-linecap="round"/>
						</svg>
					</div>
					<span class="text-zinc-400 text-sm font-medium">Today</span>
				</div>
				<div class="flex items-end justify-between">
					<span class="text-4xl font-bold text-white tabular-nums">{stats.today}</span>
					<span class="text-green-400 text-sm font-medium">+{stats.today}</span>
				</div>
			</div>

			<!-- Average Score -->
			<div class="bg-seer-surface border border-seer-border rounded-lg p-6 card-hover">
				<div class="flex items-center gap-3 mb-3">
					<div class="p-2 rounded-lg bg-purple-500/10">
						<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path d="M23 6l-9.5 9.5-5-5L1 18" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
							<path d="M17 6h6v6" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</div>
					<span class="text-zinc-400 text-sm font-medium">Average Score</span>
				</div>
				<div class="flex items-end justify-between">
					<span class="text-4xl font-bold text-white tabular-nums">{filteredAverageScore()}</span>
					<span class="text-zinc-500 text-sm font-medium">/100</span>
				</div>
			</div>

			<!-- Sources Active -->
			<div class="bg-seer-surface border border-seer-border rounded-lg p-6 card-hover">
				<div class="flex items-center justify-between mb-3">
					<div class="flex items-center gap-3">
						<div class="p-2 rounded-lg bg-purple-500/10">
							<svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path d="M22 12h-4l-3 9L9 3l-3 9H2" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
							</svg>
						</div>
						<span class="text-zinc-400 text-sm font-medium">Sources Active</span>
					</div>
					<button
						type="button"
						class="p-1.5 rounded-lg bg-seer-elevated border border-seer-border hover:border-purple-500/50 hover:bg-purple-500/10 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						onclick={handleRefetch}
						disabled={refetching}
						title="Refetch opportunities from all sources"
					>
						<svg class="w-4 h-4 text-zinc-400 hover:text-purple-400 {refetching ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</button>
				</div>
				<div class="flex items-end justify-between">
					<span class="text-4xl font-bold text-white tabular-nums">{enabledSourcesCount()}</span>
					{#if refetching}
						<span class="text-purple-400 text-sm font-medium flex items-center gap-1">
							<span class="w-2 h-2 bg-purple-400 rounded-full animate-pulse"></span>
							Fetching...
						</span>
					{:else if enabledSourcesCount() > 0}
						<span class="text-green-400 text-sm font-medium flex items-center gap-1">
							<span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
							Live
						</span>
					{:else}
						<span class="text-red-400 text-sm font-medium flex items-center gap-1">
							<span class="w-2 h-2 bg-red-400 rounded-full"></span>
							Offline
						</span>
					{/if}
				</div>
			</div>
		</div>
	{/if}

	<!-- Filters - EXACT seer-ui -->
	<div class="bg-seer-surface border border-seer-border rounded-lg p-4 mb-6">
		<div class="flex flex-wrap items-center gap-6">
			<!-- Filter icon + label -->
			<div class="flex items-center gap-2 text-zinc-400">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path d="M22 3H2l8 9.46V19l4 2v-8.54L22 3z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
				</svg>
				<span class="text-sm font-medium">Filters</span>
			</div>

			<!-- Source Dropdown -->
			<div class="flex items-center gap-2">
				<label for="source-filter" class="text-zinc-500 text-sm">Source</label>
				<select
					id="source-filter"
					class="bg-seer-elevated border border-seer-border rounded-lg px-3 py-2 text-sm text-white cursor-pointer focus:outline-none focus:border-purple-500 focus:ring-2 focus:ring-purple-500/20"
					bind:value={selectedSource}
					onchange={loadData}
				>
					<option value="">All Sources</option>
					{#if stats}
						{#each Object.keys(stats.by_source) as source}
							<option value={source}>{source}</option>
						{/each}
					{/if}
				</select>
			</div>

			<!-- Score Slider -->
			<div class="flex items-center gap-3 flex-1 max-w-xs">
				<label for="score-filter" class="text-zinc-500 text-sm whitespace-nowrap">Min Score</label>
				<input
					id="score-filter"
					type="range"
					min="0"
					max="100"
					bind:value={minScore}
					onchange={loadData}
					class="flex-1"
				/>
				<span class="text-white text-sm font-medium w-8 text-right">{minScore}</span>
			</div>

			<!-- Search -->
			<div class="flex items-center gap-2 flex-1 max-w-xs ml-auto">
				<div class="relative flex-1">
					<svg class="w-4 h-4 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<circle cx="11" cy="11" r="8" stroke-width="2"/>
						<path d="M21 21l-4.35-4.35" stroke-width="2" stroke-linecap="round"/>
					</svg>
					<input
						type="text"
						placeholder="Search opportunities..."
						class="w-full bg-seer-elevated border border-seer-border rounded-lg pl-10 pr-4 py-2 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-purple-500 focus:ring-2 focus:ring-purple-500/20"
						bind:value={searchQuery}
					/>
				</div>
			</div>
		</div>
	</div>

	<!-- Opportunities Table - EXACT seer-ui Tailwind classes -->
	<div class="bg-seer-surface border border-seer-border rounded-lg overflow-hidden">
		<!-- Selection Bar -->
		{#if selectedCount() > 0}
			<div class="flex items-center justify-between px-6 py-3 bg-purple-500/10 border-b border-purple-500/30">
				<div class="flex items-center gap-3">
					<span class="text-purple-400 text-sm font-medium">{selectedCount()} opportunit{selectedCount() === 1 ? 'y' : 'ies'} selected</span>
					<button
						type="button"
						class="text-zinc-400 text-xs hover:text-white hover:underline transition-colors cursor-pointer"
						onclick={() => analysisSet = new Set()}
					>
						Clear all
					</button>
				</div>
				<button
					type="button"
					class="btn btn-primary btn-sm"
					onclick={generatePrompt}
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path d="M13 10V3L4 14h7v7l9-11h-7z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
					</svg>
					Generate AI Prompt
				</button>
			</div>
		{/if}

		<!-- Table Header -->
		<div class="grid grid-cols-12 gap-4 px-6 py-3 border-b border-seer-border bg-seer-elevated/50 text-zinc-400 text-sm font-medium">
			<div class="col-span-1"></div>
			<button class="col-span-1 text-left hover:text-purple-400 transition-colors" onclick={() => handleSort('score')}>
				Score{getSortIndicator('score')}
			</button>
			<button class="col-span-4 text-left hover:text-purple-400 transition-colors" onclick={() => handleSort('title')}>
				Title{getSortIndicator('title')}
			</button>
			<button class="col-span-2 text-left hover:text-purple-400 transition-colors" onclick={() => handleSort('source_type')}>
				Source{getSortIndicator('source_type')}
			</button>
			<div class="col-span-1">Signals</div>
			<button class="col-span-2 text-left hover:text-purple-400 transition-colors" onclick={() => handleSort('detected_at')}>
				Detected{getSortIndicator('detected_at')}
			</button>
			<div class="col-span-1"></div>
		</div>

		{#if loading}
			<div class="flex justify-center items-center p-12">
				<div class="w-8 h-8 border-2 border-seer-border border-t-purple-500 rounded-full animate-spin"></div>
			</div>
		{:else if opportunities.length === 0}
			<div class="text-center p-12 text-zinc-500">
				No opportunities found. Check your sources or adjust filters.
			</div>
		{:else}
			<!-- Table Rows -->
			<div class="divide-y divide-seer-border">
				{#each paginatedOpportunities() as opp}
					<div class="grid grid-cols-12 gap-4 px-6 py-4 items-center hover:bg-seer-elevated/50 transition-colors">
						<!-- Checkbox -->
						<div class="col-span-1">
							<button
								type="button"
								class="p-1"
								onclick={(e) => { e.stopPropagation(); toggleAnalysis(opp.id); }}
								aria-label={isInAnalysis(opp.id) ? 'Remove from analysis' : 'Add to analysis'}
							>
								<div class="w-5 h-5 rounded border-2 flex items-center justify-center transition-colors {isInAnalysis(opp.id) ? 'bg-purple-500 border-purple-500' : 'border-zinc-500 bg-seer-bg hover:border-purple-400'}">
									{#if isInAnalysis(opp.id)}
										<svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="3">
											<path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round"/>
										</svg>
									{/if}
								</div>
							</button>
						</div>
						<!-- Score -->
						<div class="col-span-1">
							<span class="{getScoreClass(opp.score)} inline-flex items-center justify-center w-10 h-10 rounded-lg text-sm font-bold">{opp.score}</span>
						</div>
						<!-- Title -->
						<button
							type="button"
							class="col-span-4 text-left cursor-pointer"
							onclick={() => openDetail(opp)}
						>
							<h3 class="text-white font-medium text-sm mb-1 line-clamp-1 hover:text-purple-400 transition-colors">{opp.title}</h3>
							{#if opp.description}
								<p class="text-zinc-500 text-xs line-clamp-1">{opp.description}</p>
							{/if}
						</button>
						<!-- Source -->
						<div class="col-span-2">
							<span class="source-badge">
								<span class="w-1.5 h-1.5 bg-purple-400 rounded-full"></span>
								{opp.source_type}
							</span>
						</div>
						<!-- Signals -->
						<div class="col-span-1">
							<span class="text-zinc-400 text-sm">{opp.signals.length}</span>
						</div>
						<!-- Detected -->
						<div class="col-span-2">
							<span class="text-zinc-500 text-sm">{formatDate(opp.detected_at)}</span>
						</div>
						<!-- Details link -->
						<div class="col-span-1 text-right">
							<button
								type="button"
								class="text-purple-400 hover:text-purple-300 text-sm font-medium inline-flex items-center gap-1 transition-colors"
								onclick={() => openDetail(opp)}
							>
								Details
								<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path d="M9 5l7 7-7 7" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
								</svg>
							</button>
						</div>
					</div>
				{/each}
			</div>

			<!-- Pagination -->
			<div class="flex items-center justify-between px-6 py-4 border-t border-seer-border bg-seer-elevated/30">
				<span class="text-zinc-500 text-sm">
					Showing {(currentPage - 1) * pageSize + 1}-{Math.min(currentPage * pageSize, filteredTotal())} of {filteredTotal()} opportunities
					{#if searchQuery}
						<span class="text-zinc-600">(filtered from {totalItems})</span>
					{/if}
				</span>

				<select
					id="page-size"
					class="bg-seer-elevated border border-seer-border rounded-lg px-3 py-1.5 text-sm text-white cursor-pointer focus:outline-none focus:border-purple-500"
					bind:value={pageSize}
					onchange={handlePageSizeChange}
				>
					<option value={5}>5</option>
					<option value={10}>10</option>
					<option value={20}>20</option>
					<option value={50}>50</option>
				</select>

				<div class="flex items-center gap-2">
					<button
						class="px-3 py-1.5 rounded-lg bg-seer-elevated border border-seer-border text-zinc-400 text-sm hover:text-white hover:border-seer-border-hover transition-colors disabled:opacity-50"
						disabled={currentPage === 1}
						onclick={() => goToPage(currentPage - 1)}
					>
						Previous
					</button>

					{#each getPageNumbers() as page}
						{#if page === -1}
							<span class="text-zinc-500">...</span>
						{:else}
							<button
								class="px-3 py-1.5 rounded-lg text-sm transition-colors {page === currentPage ? 'bg-purple-500/10 border border-purple-500/30 text-purple-400 font-medium' : 'bg-seer-elevated border border-seer-border text-zinc-400 hover:text-white hover:border-seer-border-hover'}"
								onclick={() => goToPage(page)}
							>
								{page}
							</button>
						{/if}
					{/each}

					<button
						class="px-3 py-1.5 rounded-lg bg-seer-elevated border border-seer-border text-zinc-400 text-sm hover:text-white hover:border-seer-border-hover transition-colors disabled:opacity-50"
						disabled={currentPage === totalPages()}
						onclick={() => goToPage(currentPage + 1)}
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Detail Modal (daisyUI) -->
<dialog bind:this={detailModal} class="modal">
	<div class="modal-box bg-seer-surface border border-seer-border max-w-2xl p-0">
		{#if selectedOpportunity}
			<!-- Modal Header -->
			<div class="flex items-start justify-between p-6 border-b border-seer-border">
				<div class="flex items-center gap-4">
					<span class="{getScoreClass(selectedOpportunity.score)} inline-flex items-center justify-center w-14 h-14 rounded-xl text-xl font-bold">
						{selectedOpportunity.score}
					</span>
					<div>
						<span class="source-badge mb-2">
							<span class="w-1.5 h-1.5 bg-purple-400 rounded-full"></span>
							{selectedOpportunity.source_type}
						</span>
						<p class="text-zinc-500 text-sm">{formatDate(selectedOpportunity.detected_at)}</p>
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
			<div class="p-6 space-y-6 max-h-[60vh] overflow-y-auto">
				<!-- Title -->
				<div>
					<h2 class="text-xl font-semibold text-white mb-2">{selectedOpportunity?.title}</h2>
				</div>

				<!-- Description -->
				{#if selectedOpportunity.description}
					<div>
						<h3 class="text-sm font-medium text-zinc-400 mb-2">Description</h3>
						<p class="text-zinc-300 text-sm leading-relaxed whitespace-pre-wrap">{selectedOpportunity.description}</p>
					</div>
				{/if}

				<!-- Signals Detected -->
				<div>
					<h3 class="text-sm font-medium text-zinc-400 mb-3">Signals Detected</h3>
					<div class="flex flex-wrap gap-2">
						{#each selectedOpportunity.signals as signal}
							<span class="signal-badge inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium">
								<svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24">
									<path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z"/>
								</svg>
								{formatSignal(signal)}
							</span>
						{/each}
					</div>
				</div>

				<!-- Score Breakdown -->
				<div>
					<h3 class="text-sm font-medium text-zinc-400 mb-3">Score Breakdown</h3>
					<div class="bg-seer-elevated rounded-lg p-4 space-y-3">
						{#each getScoreBreakdown(selectedOpportunity.signals) as item}
							<div class="flex items-center justify-between">
								<span class="text-zinc-300 text-sm">{formatSignal(item.signal)}</span>
								<div class="flex items-center gap-2">
									<div class="w-24 h-2 bg-seer-bg rounded-full overflow-hidden">
										<div
											class="h-full bg-purple-500 rounded-full"
											style="width: {Math.min(item.contribution * 4, 100)}%"
										></div>
									</div>
									<span class="text-purple-400 text-sm font-medium w-8 text-right">+{item.contribution}</span>
								</div>
							</div>
						{/each}
						<div class="flex items-center justify-between pt-3 border-t border-seer-border">
							<span class="text-white font-medium">Total Score</span>
							<span class="{getScoreClass(selectedOpportunity.score)} px-3 py-1 rounded-lg text-sm font-bold">
								{selectedOpportunity.score}
							</span>
						</div>
					</div>
				</div>

				<!-- Add to Analysis Checkbox -->
				{#if selectedOpportunity}
					{@const oppId = selectedOpportunity.id}
					<div class="flex items-center gap-3 p-4 bg-seer-elevated rounded-lg">
						<button
							type="button"
							onclick={() => toggleAnalysis(oppId)}
							class="flex items-center gap-3 cursor-pointer w-full text-left"
						>
							<div class="w-5 h-5 rounded border-2 flex items-center justify-center transition-colors {isInAnalysis(oppId) ? 'bg-purple-500 border-purple-500' : 'border-zinc-500 bg-seer-bg'}">
								{#if isInAnalysis(oppId)}
									<svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="3">
										<path d="M5 13l4 4L19 7" stroke-linecap="round" stroke-linejoin="round"/>
									</svg>
								{/if}
							</div>
							<div>
								<span class="text-white font-medium">Add to Analysis</span>
								<p class="text-zinc-500 text-xs">Include this opportunity in your next report</p>
							</div>
						</button>
						{#if isInAnalysis(oppId)}
							<span class="text-green-400 text-xs font-medium whitespace-nowrap">Added</span>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Modal Footer -->
			<div class="flex items-center justify-between p-6 border-t border-seer-border bg-seer-elevated/30">
				<form method="dialog">
					<button class="btn btn-ghost btn-sm">Close</button>
				</form>
				<a
					href={selectedOpportunity.source_url}
					target="_blank"
					rel="noopener noreferrer"
					class="btn btn-primary btn-sm"
				>
					View Original
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6M15 3h6v6M10 14L21 3" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
					</svg>
				</a>
			</div>
		{/if}
	</div>
	<form method="dialog" class="modal-backdrop">
		<button>close</button>
	</form>
</dialog>

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
					<p class="text-zinc-500 text-sm">{selectedCount()} opportunities included</p>
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
				<pre class="text-zinc-300 text-sm whitespace-pre-wrap font-mono">{generatedPrompt}</pre>
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
					onclick={copyPrompt}
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
	/* Score badge colors - EXACT seer-ui */
	.score-low {
		background-color: #27272A;
		color: #71717A;
	}

	.score-mid {
		background-color: rgba(251, 191, 36, 0.15);
		color: #FBBF24;
		border: 1px solid rgba(251, 191, 36, 0.3);
	}

	.score-high {
		background-color: rgba(34, 197, 94, 0.15);
		color: #22C55E;
		border: 1px solid rgba(34, 197, 94, 0.3);
		box-shadow: 0 0 15px rgba(34, 197, 94, 0.3);
	}

	/* Signal badge for modal */
	.signal-badge {
		background-color: rgba(139, 92, 246, 0.15);
		border: 1px solid rgba(139, 92, 246, 0.25);
		color: #A78BFA;
	}

	/* Source badge - purple tag style */
	.source-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.25rem 0.625rem;
		background-color: rgba(139, 92, 246, 0.1);
		border: 1px solid rgba(139, 92, 246, 0.3);
		border-radius: 9999px;
		color: #A78BFA;
		font-size: 0.75rem;
		font-weight: 500;
	}
</style>
