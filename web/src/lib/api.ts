const API_BASE = '/api';

export interface Opportunity {
	id: number;
	title: string;
	description: string;
	source_type: string;
	source_url: string;
	source_id_external: string;
	score: number;
	signals: string[];
	detected_at: string;
	created_at: string;
}

export interface Source {
	id: number;
	type: string;
	name: string;
	url?: string;
	enabled: boolean;
	is_builtin: boolean;
	created_at: string;
}

export interface SourceTypes {
	types: string[];
	is_pro: boolean;
	max_rss: number;
}

export interface Stats {
	total: number;
	by_source: Record<string, number>;
	average_score: number;
	today: number;
}

export interface Report {
	id: number;
	period_start: string;
	period_end: string;
	opportunity_count: number;
	content_human?: string;
	content_prompt?: string;
	summary?: string;
	ai_analysis?: string;
	generated_at: string;
	created_at: string;
}

// Opportunities
export async function getOpportunities(params?: {
	source?: string;
	min_score?: number;
	limit?: number;
	offset?: number;
}): Promise<Opportunity[]> {
	const searchParams = new URLSearchParams();
	if (params?.source) searchParams.set('source', params.source);
	if (params?.min_score) searchParams.set('min_score', params.min_score.toString());
	if (params?.limit) searchParams.set('limit', params.limit.toString());
	if (params?.offset) searchParams.set('offset', params.offset.toString());

	const query = searchParams.toString();
	const url = `${API_BASE}/opportunities${query ? `?${query}` : ''}`;
	const res = await fetch(url);
	return res.json();
}

export async function getOpportunity(id: number): Promise<Opportunity> {
	const res = await fetch(`${API_BASE}/opportunities/${id}`);
	return res.json();
}

export async function getStats(): Promise<Stats> {
	const res = await fetch(`${API_BASE}/opportunities/stats`);
	return res.json();
}

// Sources
export async function getSources(): Promise<Source[]> {
	const res = await fetch(`${API_BASE}/sources`);
	return res.json();
}

export async function getSourceTypes(): Promise<SourceTypes> {
	const res = await fetch(`${API_BASE}/sources/types`);
	return res.json();
}

export async function createSource(data: { type: string; name: string; url?: string }): Promise<Source> {
	const res = await fetch(`${API_BASE}/sources`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	return res.json();
}

export async function toggleSource(id: number): Promise<Source> {
	const res = await fetch(`${API_BASE}/sources/${id}/toggle`, { method: 'POST' });
	return res.json();
}

export async function deleteSource(id: number): Promise<void> {
	await fetch(`${API_BASE}/sources/${id}`, { method: 'DELETE' });
}

// Reports
export async function getReports(): Promise<Report[]> {
	const res = await fetch(`${API_BASE}/reports`);
	return res.json();
}

export async function generateReport(start?: string, end?: string): Promise<Report> {
	const searchParams = new URLSearchParams();
	if (start) searchParams.set('start', start);
	if (end) searchParams.set('end', end);

	const query = searchParams.toString();
	const url = `${API_BASE}/reports/generate${query ? `?${query}` : ''}`;
	const res = await fetch(url, { method: 'POST' });
	return res.json();
}

export async function getReport(id: number): Promise<Report> {
	const res = await fetch(`${API_BASE}/reports/${id}`);
	return res.json();
}

export async function getReportPrompt(id: number): Promise<string> {
	const res = await fetch(`${API_BASE}/reports/${id}/prompt`);
	return res.text();
}
