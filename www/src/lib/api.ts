import decode from 'jwt-decode';
import { API_BASE_URL } from './env';

export async function createTable(name: string): Promise<[string, { aud: string; tbl: string }]> {
	const res = await fetch(`${API_BASE_URL}/tables`, {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify({ name }),
	});

	if (res.status !== 201) {
		throw new Error(`unexpected status code ${res.status}`);
	}

	const { token } = await res.json();
	return [token, decode(token)];
}
