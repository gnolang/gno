// In-page wallet discovery — the browser-extension half of GnoConnect.
//
// A wallet that runs code in the page announces itself by dispatching
// "gno:registerWallet" on window; the page asks for announcements by
// dispatching "gno:requestWallet". Neither side can assume it loaded first,
// so wallets announce on load AND on every request, and the page keeps
// listening after it has asked. Same handshake as EIP-6963 / the Wallet
// Standard, in the vocabulary of the GnoConnect draft (registerWallet /
// getWallets).
//
// External wallets — mobile apps, standalone signers — cannot announce on
// window at all. They are discovered through the server-embedded registry
// (components/wallets.json) and reached by launch link; the chooser merges
// both kinds.

export const REGISTER_WALLET_EVENT = "gno:registerWallet";
export const REQUEST_WALLET_EVENT = "gno:requestWallet";

// Identity of an announced wallet. `uuid` is per announcement (a page-load
// handle); `rdns` is the durable identity, reverse-DNS of the wallet vendor.
export interface GnoWalletInfo {
	uuid: string;
	name: string;
	icon: string; // data:image/ URI
	rdns: string;
}

// A rejection is a normal outcome, not an error: the user declining is an
// answer, and only a real failure throws.
export type UserResponse<T> =
	| { status: "Approved"; args: T }
	| { status: "Rejected" };

// The transaction intent, field for field the `tx` launch link an external
// wallet receives — named args included, so the two transports carry the same
// payload and a realm parameter cannot collide with a reserved key.
export interface GnoTxRequest {
	path: string;
	func: string;
	args: { name: string; value: string }[];
	send?: string;
	rpc?: string;
	chainid?: string;
}

// Only the method gnoweb calls is declared. It is optional: a wallet may
// announce itself while implementing more (or less) of the draft's surface,
// and the caller must degrade instead of assuming.
export interface GnoWalletProvider {
	signAndSubmitTransaction?(
		tx: GnoTxRequest,
	): Promise<UserResponse<{ hash: string }>>;
}

export interface GnoWallet {
	info: GnoWalletInfo;
	provider: GnoWalletProvider;
}

// Anything running in the page can announce a wallet, so the announcement is
// untrusted input: cap the list so a hostile script cannot push the real
// wallet out of the chooser, and clamp the name it renders.
const MAX_WALLETS = 16;
const MAX_NAME_LENGTH = 64;
// Every controller instance on the page asks at connect time; one dispatch
// covers them all.
const REQUEST_COALESCE_MS = 250;

const wallets = new Map<string, GnoWallet>(); // by info.uuid
const listeners = new Set<(wallets: GnoWallet[]) => void>();
let lastRequest = 0;
let capacityWarned = false;

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

// Accept an announcement only if it can be rendered and called safely. A
// non-conforming icon is dropped rather than rejecting the wallet: an
// unreachable wallet is a worse outcome than a missing image.
function normalize(detail: unknown): GnoWallet | null {
	if (!isRecord(detail)) return null;
	const { info, provider } = detail;
	if (!isRecord(info) || !isRecord(provider)) return null;

	const uuid = typeof info.uuid === "string" ? info.uuid.trim() : "";
	const name = typeof info.name === "string" ? info.name.trim() : "";
	const rdns = typeof info.rdns === "string" ? info.rdns.trim() : "";
	if (!uuid || !name) return null;

	const icon =
		typeof info.icon === "string" && info.icon.startsWith("data:image/")
			? info.icon
			: "";

	return {
		info: { uuid, name: name.slice(0, MAX_NAME_LENGTH), icon, rdns },
		provider: provider as GnoWalletProvider,
	};
}

function onRegister(event: Event): void {
	const wallet = normalize((event as CustomEvent).detail);
	if (!wallet) {
		console.warn("gno:registerWallet: malformed announcement, ignored");
		return;
	}
	if (!wallets.has(wallet.info.uuid) && wallets.size >= MAX_WALLETS) {
		if (!capacityWarned) {
			capacityWarned = true;
			console.warn(
				`gno:registerWallet: more than ${MAX_WALLETS} wallets announced, ignoring the rest`,
			);
		}
		return;
	}

	wallets.set(wallet.info.uuid, wallet);
	const snapshot = Array.from(wallets.values());
	listeners.forEach((listener) => {
		try {
			listener(snapshot);
		} catch (err) {
			console.warn("gno:registerWallet: listener failed", err);
		}
	});
}

// Listening starts at module evaluation, before any caller can ask: an
// announcement that lands in between would otherwise be lost, and a wallet
// that announces only once would never appear.
window.addEventListener(REGISTER_WALLET_EVENT, onRegister);

// getWallets is the page-side API of the draft: the wallets announced so far,
// a subscription for the ones still to come, and the request that prompts
// them. Announcements arrive whenever a wallet loads, so `get()` is a
// snapshot, never a final answer.
export function getWallets(): {
	get(): GnoWallet[];
	on(listener: (wallets: GnoWallet[]) => void): () => void;
	request(): void;
} {
	return {
		get: () => Array.from(wallets.values()),
		on: (listener) => {
			listeners.add(listener);
			return () => listeners.delete(listener);
		},
		request: () => {
			const now = Date.now();
			if (now - lastRequest < REQUEST_COALESCE_MS) return;
			lastRequest = now;
			window.dispatchEvent(new Event(REQUEST_WALLET_EVENT));
		},
	};
}
