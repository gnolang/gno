// The connected identity, client-side only. No cookie, no server session, no
// signature challenge: every page stays cacheable and identical for all
// visitors. What connecting buys is the header avatar, an Execute that goes
// straight to the remembered wallet, and a `signer` pin on every intent.

import {
	type GnoAccount,
	type GnoWallet,
	getWallets,
} from "../../../frontend/js/wallet-discovery.js";

export interface GnoSession {
	rdns: string; // durable wallet identity — never the per-page-load uuid
	address: string;
	chainid: string;
	name: string;
}

const STORAGE_KEY = "gnoweb:session:v1";
const MAX_NAME_LENGTH = 64;

const listeners = new Set<(session: GnoSession | null) => void>();

function emit(session: GnoSession | null): void {
	listeners.forEach((listener) => {
		try {
			listener(session);
		} catch (err) {
			console.warn("session: listener failed", err);
		}
	});
}

// Storage is same-origin, but it is still parsed input: a half-written or
// hand-edited entry must drop the session, not render undefined.
function normalize(value: unknown): GnoSession | null {
	if (typeof value !== "object" || value === null) return null;
	const { rdns, address, chainid, name } = value as Record<string, unknown>;
	if (typeof rdns !== "string" || !rdns) return null;
	if (typeof address !== "string" || !address) return null;
	return {
		rdns,
		address,
		chainid: typeof chainid === "string" ? chainid : "",
		name: typeof name === "string" ? name.slice(0, MAX_NAME_LENGTH) : "",
	};
}

export function readSession(): GnoSession | null {
	let raw: string | null = null;
	try {
		raw = window.localStorage.getItem(STORAGE_KEY);
	} catch {
		return null; // storage disabled (private mode, blocked cookies)
	}
	if (!raw) return null;
	try {
		return normalize(JSON.parse(raw));
	} catch {
		return null;
	}
}

export function writeSession(session: GnoSession): void {
	try {
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
	} catch {
		// A session we cannot persist is still usable for this page load.
	}
	emit(session);
}

export function clearSession(): void {
	try {
		window.localStorage.removeItem(STORAGE_KEY);
	} catch {
		// Nothing to do: the in-memory notification below is what the UI reads.
	}
	emit(null);
}

export function onSessionChange(
	fn: (session: GnoSession | null) => void,
): () => void {
	listeners.add(fn);
	return () => listeners.delete(fn);
}

export function findAnnounced(rdns: string): GnoWallet | undefined {
	return getWallets()
		.get()
		.find((wallet) => wallet.info.rdns === rdns);
}

function toSession(wallet: GnoWallet, account: GnoAccount): GnoSession {
	return {
		rdns: wallet.info.rdns,
		address: account.address,
		chainid: account.chainid ?? "",
		name: wallet.info.name,
	};
}

// Ask a wallet who the user is and remember the answer. A wallet that declines,
// throws, or does not implement connect yields null — the caller degrades, it
// never dead-ends.
export async function connectWallet(
	wallet: GnoWallet,
): Promise<GnoSession | null> {
	const connect = wallet.provider?.connect;
	if (typeof connect !== "function") {
		console.warn(
			`session: wallet "${wallet.info.name}" announced no connect; it cannot be logged into`,
		);
		return null;
	}
	try {
		const response = await connect.call(wallet.provider);
		if (response?.status !== "Approved" || !response.args?.address) return null;
		const session = toSession(wallet, response.args);
		writeSession(session);
		return session;
	} catch (err) {
		console.warn(
			`session: wallet "${wallet.info.name}" failed to connect`,
			err,
		);
		return null;
	}
}

// Re-check the remembered address against the wallet, without prompting.
// getAccount first (it is defined never to prompt); connect second, which per
// the standard must resolve silently for an approved origin. A different
// address updates the session; a failure drops it.
export async function reconcile(): Promise<GnoSession | null> {
	const session = readSession();
	if (!session) return null;

	const wallet = findAnnounced(session.rdns);
	if (!wallet) return session; // not announced yet — keep what we render

	const ask = async (): Promise<GnoAccount | null> => {
		const getAccount = wallet.provider?.getAccount;
		if (typeof getAccount === "function") {
			const response = await getAccount.call(wallet.provider);
			if (response?.status === "Approved") return response.args;
		}
		const connect = wallet.provider?.connect;
		if (typeof connect === "function") {
			const response = await connect.call(wallet.provider);
			if (response?.status === "Approved") return response.args;
		}
		return null;
	};

	let account: GnoAccount | null = null;
	try {
		account = await ask();
	} catch (err) {
		console.warn("session: reconcile failed", err);
	}
	if (!account?.address) {
		clearSession();
		return null;
	}
	if (
		account.address === session.address &&
		account.chainid === session.chainid
	) {
		return session;
	}
	const updated = toSession(wallet, account);
	writeSession(updated);
	return updated;
}

// Header display form: enough of each end to recognise, short enough to fit.
export function truncate(address: string): string {
	return address.length <= 15
		? address
		: `${address.slice(0, 8)}…${address.slice(-4)}`;
}
