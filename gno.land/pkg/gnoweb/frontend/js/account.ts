import { abciQuery, ChainError, qevalJSON, readMeta } from "./chain.js";

// Per-address chain data for the connected account, cached for the page and
// shared by every controller bundle through a globalThis slot.

export interface Account {
	address: string;
	accountNumber: string; // decimal strings: uint64 exceeds Number
	sequence: string;
	ugnot: bigint;
	otherDenoms: number; // counted only; denom strings are never displayed
}

const ADDRESS = /^g1[02-9ac-hj-np-z]{38}$/;
const COIN = /^(\d+)([a-z][a-z0-9/:._-]{1,127})$/;
const CACHE = Symbol.for("gnoweb.account.cache");

export function isAddress(s: string): boolean {
	return ADDRESS.test(s);
}

function cache(): Map<string, Promise<unknown>> {
	const slot = globalThis as unknown as Record<
		symbol,
		Map<string, Promise<unknown>> | undefined
	>;
	slot[CACHE] ??= new Map();
	return slot[CACHE];
}

function cached<T>(key: string, load: () => Promise<T>): Promise<T> {
	const entries = cache();
	let promise = entries.get(key) as Promise<T> | undefined;
	if (!promise) {
		promise = load();
		entries.set(key, promise);
		promise.catch(() => entries.delete(key));
	}
	return promise;
}

export function refresh(address: string): void {
	cache().delete(`account:${address}`);
}

export function parseAccount(
	address: string,
	text: string | null,
): Account | null {
	if (text === null) return null;
	let base:
		| { coins?: string; account_number?: string; sequence?: string }
		| undefined;
	try {
		base = JSON.parse(text)?.BaseAccount;
	} catch {
		throw new ChainError(`auth/accounts/${address}: unreadable account`);
	}
	if (!base) return null;

	let ugnot = 0n;
	let otherDenoms = 0;
	for (const coin of (base.coins ?? "").split(",")) {
		const match = COIN.exec(coin.trim());
		if (!match) continue;
		if (match[2] === "ugnot") ugnot += BigInt(match[1]);
		else otherDenoms++;
	}
	return {
		address,
		accountNumber: String(base.account_number ?? "0"),
		sequence: String(base.sequence ?? "0"),
		ugnot,
		otherDenoms,
	};
}

// formatGnot is the ugnot amount in GNOT, without the unit.
export function formatGnot(ugnot: bigint): string {
	const whole = (ugnot / 1_000_000n).toLocaleString("en-US");
	const frac = ugnot % 1_000_000n;
	if (frac === 0n) return whole;
	return `${whole}.${frac.toString().padStart(6, "0").replace(/0+$/, "")}`;
}

export interface AccountView {
	amount: string; // GNOT, without the unit
	extra: string | null; // other tokens, or "not on chain yet"
	accountNumber: string | null;
	sequence: string | null;
}

export function accountView(account: Account | null): AccountView {
	if (!account) {
		return {
			amount: "0",
			extra: "Not on chain yet",
			accountNumber: null,
			sequence: null,
		};
	}
	const n = account.otherDenoms;
	return {
		amount: formatGnot(account.ugnot),
		extra: n === 0 ? null : `+${n} other token${n === 1 ? "" : "s"}`,
		accountNumber: account.accountNumber,
		sequence: account.sequence,
	};
}

export function commandNumbers(account: Account | null): {
	accountNumber: string;
	sequence: string;
} {
	return account
		? { accountNumber: account.accountNumber, sequence: account.sequence }
		: { accountNumber: "ACCOUNTNUMBER", sequence: "SEQUENCENUMBER" };
}

// chainMatches is false only when the session and the page name different chains.
export function chainMatches(sessionChainid: string): boolean {
	const page = readMeta("gnoconnect:chainid");
	return !sessionChainid || !page || sessionChainid === page;
}

// forAddress caches load under kind:address; a non-address sends no request.
function forAddress<T>(
	kind: string,
	address: string,
	load: () => Promise<T>,
): Promise<T> {
	if (!isAddress(address)) {
		return Promise.reject(new ChainError("not an address"));
	}
	return cached(`${kind}:${address}`, load);
}

export function loadAccount(address: string): Promise<Account | null> {
	return forAddress("account", address, async () =>
		parseAccount(address, await abciQuery(`auth/accounts/${address}`)),
	);
}

// loadUsername is null for an unregistered address; an RPC failure throws.
export function loadUsername(address: string): Promise<string | null> {
	return forAddress("username", address, async () => {
		try {
			const name = await qevalJSON(
				`gno.land/r/sys/users.ResolveAddress("${address}").Name()`,
			);
			return name || null;
		} catch (err) {
			// The node evaluated and failed: a nil *UserData, i.e. not registered.
			if (err instanceof ChainError && err.nodeError !== undefined) return null;
			throw err;
		}
	});
}
