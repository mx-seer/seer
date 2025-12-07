<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getSources,
		getSourceTypes,
		createSource,
		toggleSource,
		deleteSource,
		type Source,
		type SourceTypes
	} from '$lib/api';

	let sources: Source[] = $state([]);
	let sourceTypes: SourceTypes | null = $state(null);
	let loading = $state(true);
	let showAddModal = $state(false);

	// Form state
	let newSourceType = $state('');
	let newSourceName = $state('');
	let newSourceUrl = $state('');
	let formError = $state('');

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		try {
			const [s, types] = await Promise.all([getSources(), getSourceTypes()]);
			sources = s;
			sourceTypes = types;
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

	async function handleDelete(id: number) {
		if (!confirm('Are you sure you want to delete this source?')) return;
		try {
			await deleteSource(id);
			sources = sources.filter((s) => s.id !== id);
		} catch (e) {
			console.error('Failed to delete source:', e);
		}
	}

	async function handleAdd() {
		formError = '';

		if (!newSourceType) {
			formError = 'Please select a source type';
			return;
		}
		if (!newSourceName) {
			formError = 'Please enter a name';
			return;
		}
		if (newSourceType === 'rss' && !newSourceUrl) {
			formError = 'RSS sources require a URL';
			return;
		}

		try {
			const created = await createSource({
				type: newSourceType,
				name: newSourceName,
				url: newSourceUrl || undefined
			});
			sources = [...sources, created];
			showAddModal = false;
			resetForm();
		} catch (e) {
			formError = 'Failed to create source. You may have reached the RSS limit.';
		}
	}

	function resetForm() {
		newSourceType = '';
		newSourceName = '';
		newSourceUrl = '';
		formError = '';
	}

	function openAddModal() {
		resetForm();
		showAddModal = true;
	}
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-2xl font-bold">Sources</h1>
		<button class="btn btn-primary" onclick={openAddModal}>Add Source</button>
	</div>

	{#if sourceTypes}
		<div class="alert alert-info">
			<span>
				{#if sourceTypes.is_pro}
					Pro Edition - All features unlocked
				{:else}
					Community Edition - RSS feeds limited to {sourceTypes.max_rss}
				{/if}
			</span>
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center p-8">
			<span class="loading loading-spinner loading-lg"></span>
		</div>
	{:else}
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each sources as source}
				<div class="card bg-base-100 shadow-xl">
					<div class="card-body">
						<div class="flex justify-between items-start">
							<div>
								<h2 class="card-title">{source.name}</h2>
								<div class="badge badge-outline">{source.type}</div>
								{#if source.is_builtin}
									<div class="badge badge-primary badge-sm ml-1">Built-in</div>
								{/if}
							</div>
							<input
								type="checkbox"
								class="toggle toggle-primary"
								checked={source.enabled}
								onchange={() => handleToggle(source.id)}
							/>
						</div>

						{#if source.url}
							<p class="text-sm text-base-content/60 truncate" title={source.url}>
								{source.url}
							</p>
						{/if}

						<div class="card-actions justify-end mt-4">
							{#if !source.is_builtin}
								<button class="btn btn-error btn-sm" onclick={() => handleDelete(source.id)}>
									Delete
								</button>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Add Source Modal -->
{#if showAddModal}
	<div class="modal modal-open">
		<div class="modal-box">
			<h3 class="font-bold text-lg">Add New Source</h3>

			{#if formError}
				<div class="alert alert-error mt-4">
					<span>{formError}</span>
				</div>
			{/if}

			<div class="form-control mt-4">
				<label class="label" for="source-type">
					<span class="label-text">Type</span>
				</label>
				<select id="source-type" class="select select-bordered" bind:value={newSourceType}>
					<option value="">Select type...</option>
					{#if sourceTypes}
						{#each sourceTypes.types as type}
							<option value={type}>{type}</option>
						{/each}
					{/if}
				</select>
			</div>

			<div class="form-control mt-4">
				<label class="label" for="source-name">
					<span class="label-text">Name</span>
				</label>
				<input
					id="source-name"
					type="text"
					class="input input-bordered"
					placeholder="My Custom Source"
					bind:value={newSourceName}
				/>
			</div>

			{#if newSourceType === 'rss'}
				<div class="form-control mt-4">
					<label class="label" for="source-url">
						<span class="label-text">URL</span>
					</label>
					<input
						id="source-url"
						type="url"
						class="input input-bordered"
						placeholder="https://example.com/feed.xml"
						bind:value={newSourceUrl}
					/>
				</div>
			{/if}

			<div class="modal-action">
				<button class="btn" onclick={() => (showAddModal = false)}>Cancel</button>
				<button class="btn btn-primary" onclick={handleAdd}>Add</button>
			</div>
		</div>
		<button class="modal-backdrop" onclick={() => (showAddModal = false)} aria-label="Close modal"></button>
	</div>
{/if}
