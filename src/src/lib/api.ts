// API configuration and client

const API_BASE = import.meta.env.PROD
	? '/api' // Production: relative path (same domain on Vercel)
	: 'http://localhost:8080/api'; // Development: local Go server

export interface VerifyNameRequest {
	name: string;
}

export interface VerifyNameResponse {
	success: boolean;
	message?: string;
}

export interface RSVPRequest {
	name: string;
	email: string;
	isAttending: boolean;
	attendingGuests: string[];
	diet?: string;
}

export interface RSVPResponse {
	success: boolean;
	message: string;
}

/**
 * Verify if a name exists on the guest list
 */
export async function verifyName(name: string): Promise<VerifyNameResponse> {
	const response = await fetch(`${API_BASE}/verify-name`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ name } as VerifyNameRequest),
	});

	if (!response.ok) {
		// Handle 404 (guest not found) gracefully
		if (response.status === 404) {
			const data = await response.json();
			return data;
		}
		throw new Error(`Failed to verify name: ${response.statusText}`);
	}

	return response.json();
}

/**
 * Submit RSVP
 */
export async function submitRSVP(data: RSVPRequest): Promise<RSVPResponse> {
	const response = await fetch(`${API_BASE}/submit-rsvp`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(data),
	});

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		throw new Error(errorData.message || `Failed to submit RSVP: ${response.statusText}`);
	}

	return response.json();
}

/**
 * Health check
 */
export async function healthCheck(): Promise<{ status: string; timestamp: number; guests: number }> {
	const response = await fetch(`${API_BASE}/health`);
	
	if (!response.ok) {
		throw new Error(`Health check failed: ${response.statusText}`);
	}

	return response.json();
}
