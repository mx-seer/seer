<script lang="ts">
	import { onMount } from 'svelte';
	import { getOpportunities, getStats, type Opportunity, type Stats } from '$lib/api';

	let opportunities: Opportunity[] = $state([]);
	let stats: Stats | null = $state(null);
	let loading = $state(true);
	let selectedSource = $state('');
	let minScore = $state(0);

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		try {
			const [opps, s] = await Promise.all([
				getOpportunities({
					source: selectedSource || undefined,
					min_score: minScore || undefined,
					limit: 50
				}),
				getStats()
			]);
			opportunities = opps;
			stats = s;
		} catch (e) {
			console.error('Failed to load data:', e);
		}
		loading = false;
	}

	function getScoreColor(score: number): string {
		if (score >= 70) return 'badge-success';
		if (score >= 40) return 'badge-warning';
		return 'badge-error';
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<div class="space-y-6">
	<!-- Stats Cards -->
	{#if stats}
		<div class="stats shadow w-full">
			<div class="stat">
				<div class="stat-title">Total Opportunities</div>
				<div class="stat-value text-primary">{stats.total}</div>
			</div>
			<div class="stat">
				<div class="stat-title">Today</div>
				<div class="stat-value text-secondary">{stats.today}</div>
			</div>
			<div class="stat">
				<div class="stat-title">Average Score</div>
				<div class="stat-value">{Math.round(stats.average_score)}</div>
			</div>
			<div class="stat">
				<div class="stat-title">Sources Active</div>
				<div class="stat-value">{Object.keys(stats.by_source).length}</div>
			</div>
		</div>
	{/if}

	<!-- Filters -->
	<div class="card bg-base-100 shadow-xl">
		<div class="card-body">
			<h2 class="card-title">Filters</h2>
			<div class="flex flex-wrap gap-4">
				<div class="form-control">
					<label class="label" for="source-filter">
						<span class="label-text">Source</span>
					</label>
					<select
						id="source-filter"
						class="select select-bordered"
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
				<div class="form-control">
					<label class="label" for="score-filter">
						<span class="label-text">Min Score: {minScore}</span>
					</label>
					<input
						id="score-filter"
						type="range"
						min="0"
						max="100"
						bind:value={minScore}
						onchange={loadData}
						class="range range-primary"
					/>
				</div>
			</div>
		</div>
	</div>

	<!-- Opportunities List -->
	<div class="card bg-base-100 shadow-xl">
		<div class="card-body">
			<h2 class="card-title">Opportunities</h2>

			{#if loading}
				<div class="flex justify-center p-8">
					<span class="loading loading-spinner loading-lg"></span>
				</div>
			{:else if opportunities.length === 0}
				<div class="text-center p-8 text-base-content/60">
					No opportunities found. Check your sources or adjust filters.
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="table table-zebra">
						<thead>
							<tr>
								<th>Score</th>
								<th>Title</th>
								<th>Source</th>
								<th>Signals</th>
								<th>Detected</th>
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each opportunities as opp}
								<tr>
									<td>
										<div class="badge {getScoreColor(opp.score)}">{opp.score}</div>
									</td>
									<td>
										<div class="font-medium max-w-md truncate" title={opp.title}>
											{opp.title}
										</div>
										{#if opp.description}
											<div class="text-sm text-base-content/60 max-w-md truncate">
												{opp.description}
											</div>
										{/if}
									</td>
									<td>
										<div class="badge badge-outline">{opp.source_type}</div>
									</td>
									<td>
										<div class="flex flex-wrap gap-1">
											{#each opp.signals.slice(0, 3) as signal}
												<div class="badge badge-ghost badge-sm">{signal}</div>
											{/each}
											{#if opp.signals.length > 3}
												<div class="badge badge-ghost badge-sm">+{opp.signals.length - 3}</div>
											{/if}
										</div>
									</td>
									<td class="text-sm">{formatDate(opp.detected_at)}</td>
									<td>
										<a
											href={opp.source_url}
											target="_blank"
											rel="noopener noreferrer"
											class="btn btn-ghost btn-xs"
										>
											View →
										</a>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</div>
