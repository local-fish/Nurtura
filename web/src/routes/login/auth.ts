import { API_URL } from '$src/config';

const TOKEN_KEY = 'auth_token';

async function post(path: string, body: object) {
	const res = await fetch(`${API_URL}${path}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	const data = await res.json().catch(() => ({}));
	if (!res.ok) {
		// writeErr() message; adjust the field name if yours differs
		throw new Error(data.error ?? data.message ?? `Request failed (${res.status})`);
	}
	return data;
}

export const register = (email: string, password: string) =>
	post('/auth/register', { email, password });

export async function login(email: string, password: string) {
	const { token, expires_in } = await post('/auth/login', { email, password });
	localStorage.setItem(
		TOKEN_KEY,
		JSON.stringify({
			token,
			expiresAt: Date.now() + expires_in * 1000
		})
	);
	return token;
}

export function getToken() {
	try {
		const saved = JSON.parse(localStorage.getItem(TOKEN_KEY) || 'token not found');
		if (saved && saved.expiresAt > Date.now()) return saved.token;
	} catch {}
	localStorage.removeItem(TOKEN_KEY);
	return null;
}

export const logout = () => localStorage.removeItem(TOKEN_KEY);

export async function authFetch(path: string, options: RequestInit = {}) {
	const token = getToken();
	const res = await fetch(`${API_URL}${path}`, {
		...options,
		headers: { ...options.headers, Authorization: `Bearer ${token}` }
	});
	if (res.status === 401) logout();
	return res;
}
