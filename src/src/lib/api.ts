// API configuration and client
// When using `vercel dev`, both frontend and API are served from the same port
// so we always use relative paths
const API_BASE = '/api';

export interface VerifyNameRequest {
	name: string;
	email: string;
}

export interface FamilyMember {
	id: string;
	name: string;
}

export interface VerifyNameResponse {
	success: boolean;
	message?: string;
	familyMembers?: FamilyMember[];
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

export interface AvatarSelection {
	guestName: string;
	avatar: string;
	message: string;
}

export interface SaveAvatarsRequest {
	email: string;
	avatars: AvatarSelection[];
}

export interface SaveAvatarsResponse {
	success: boolean;
	message: string;
}

export interface GuestAvatar {
	name: string;
	avatar: string;
	message: string;
}

export interface GetAvatarsResponse {
	success: boolean;
	avatars: GuestAvatar[];
}

/**
 * Verify if a name exists on the guest list and get family members
 */
export async function verifyName(name: string, email: string): Promise<VerifyNameResponse> {
	const response = await fetch(`${API_BASE}/verify-name`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ name, email } as VerifyNameRequest),
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
 * Save avatar selections for guests
 */
export async function saveAvatars(data: SaveAvatarsRequest): Promise<SaveAvatarsResponse> {
	const response = await fetch(`${API_BASE}/save-avatars`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(data),
	});

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		throw new Error(errorData.message || `Failed to save avatars: ${response.statusText}`);
	}

	return response.json();
}

/**
 * Get all avatars
 */
export async function getAvatars(): Promise<GetAvatarsResponse> {
	const response = await fetch(`${API_BASE}/get-avatars`, {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json',
		},
	});

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		throw new Error(errorData.message || `Failed to get avatars: ${response.statusText}`);
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
