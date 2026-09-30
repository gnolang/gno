import "./dom.js";
import assert from "node:assert/strict";
import { beforeEach, test } from "node:test";
import {
	clearSession,
	displayName,
	followWallet,
	type GnoSession,
	onSessionChange,
	readSession,
	SESSION_EVENT,
	setUsername,
	writeSession,
} from "../../feature/connect/frontend/session.js";
import type { GnoProviderEvents, GnoWallet } from "../js/wallet-discovery.js";
import { installDOM } from "./dom.js";

const alice: GnoSession = {
	rdns: "test.stub",
	address: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
	chainid: "dev",
	name: "Stub",
};

beforeEach(() => {
	installDOM();
});

test("a write and a clear notify subscribers", () => {
	const seen: (GnoSession | null)[] = [];
	const off = onSessionChange((s) => seen.push(s));
	writeSession(alice);
	clearSession();
	off();
	assert.deepEqual(seen, [alice, null]);
});

test("a change announced by another bundle's copy reaches subscribers", () => {
	const seen: (GnoSession | null)[] = [];
	const off = onSessionChange((s) => seen.push(s));
	window.dispatchEvent(new CustomEvent(SESSION_EVENT, { detail: alice }));
	off();
	assert.deepEqual(seen, [alice]);
});

test("unsubscribing stops notifications", () => {
	const seen: (GnoSession | null)[] = [];
	onSessionChange((s) => seen.push(s))();
	writeSession(alice);
	assert.deepEqual(seen, []);
});

test("a throwing subscriber does not stop the others", () => {
	const seen: (GnoSession | null)[] = [];
	const offA = onSessionChange(() => {
		throw new Error("boom");
	});
	const offB = onSessionChange((s) => seen.push(s));
	writeSession(alice);
	offA();
	offB();
	assert.deepEqual(seen, [alice]);
});

test("setUsername stores, replaces and clears the name for the current address", () => {
	writeSession(alice);
	setUsername(alice.address, "moul");
	assert.equal(readSession()?.username, "moul");
	setUsername(alice.address, null);
	assert.equal(readSession()?.username, undefined);
});

test("setUsername ignores another address and an unchanged name", () => {
	writeSession({ ...alice, username: "moul" });
	const seen: unknown[] = [];
	const off = onSessionChange((s) => seen.push(s));
	setUsername("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", "other");
	setUsername(alice.address, "moul");
	off();
	assert.deepEqual(seen, []);
	assert.equal(readSession()?.username, "moul");
});

test("a stored username is clamped and must be a string", () => {
	const { storage } = installDOM();
	storage.set(
		"gnoweb:session:v1",
		JSON.stringify({ ...alice, username: "x".repeat(100) }),
	);
	assert.equal(readSession()?.username?.length, 64);
	storage.set("gnoweb:session:v1", JSON.stringify({ ...alice, username: 42 }));
	assert.equal(readSession()?.username, undefined);
});

test("displayName prefers @username over the truncated address", () => {
	assert.equal(displayName({ ...alice, username: "moul" }), "@moul");
	assert.equal(displayName(alice), "g1jg8mtu…sqf5");
});

test("a name longer than the clamp converges instead of re-announcing", () => {
	writeSession(alice);
	setUsername(alice.address, "x".repeat(100));
	const seen: unknown[] = [];
	const off = onSessionChange((s) => seen.push(s));
	setUsername(alice.address, "x".repeat(100));
	off();
	assert.deepEqual(seen, []);
});

test("a name storage cannot keep is not announced", () => {
	writeSession(alice);
	Object.assign(window.localStorage, {
		setItem: () => {
			throw new Error("QuotaExceededError");
		},
	});
	const seen: unknown[] = [];
	const off = onSessionChange((s) => seen.push(s));
	setUsername(alice.address, "moul");
	off();
	assert.deepEqual(seen, []);
});

// A wallet whose provider records subscriptions, so tests can fire its events.
function eventfulWallet(rdns = alice.rdns) {
	const handlers = new Map<string, Set<(detail: unknown) => void>>();
	const wallet: GnoWallet = {
		info: { uuid: "u1", name: "Stub", icon: "", rdns },
		provider: {
			on(event, listener) {
				const set = handlers.get(event) ?? new Set();
				set.add(listener as (detail: unknown) => void);
				handlers.set(event, set);
				return () => set.delete(listener as (detail: unknown) => void);
			},
		},
	};
	const fire = <E extends keyof GnoProviderEvents>(
		event: E,
		detail: GnoProviderEvents[E],
	) => {
		for (const handler of handlers.get(event) ?? []) handler(detail);
	};
	return {
		wallet,
		fire,
		count: (event: string) => handlers.get(event)?.size ?? 0,
	};
}

const bob = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq";

// Follow a wallet that is already announced.
const follow = (wallet: GnoWallet) =>
	followWallet(
		() => wallet,
		() => () => {},
	);

test("followWallet rewrites the session on an account switch", () => {
	writeSession({ ...alice, username: "alice" });
	const { wallet, fire } = eventfulWallet();
	const stop = follow(wallet);
	fire("accountChanged", { address: bob, chainid: "dev", pubkey: null });
	stop();
	const session = readSession();
	assert.equal(session?.address, bob);
	assert.equal(session?.username, undefined); // the old name belongs to alice
});

test("followWallet records a network change", () => {
	writeSession(alice);
	const { wallet, fire } = eventfulWallet();
	const stop = follow(wallet);
	fire("networkChanged", {
		chainid: "test5",
		rpc: "https://rpc.test5.gno.land",
	});
	stop();
	assert.equal(readSession()?.chainid, "test5");
	assert.equal(readSession()?.address, alice.address);
});

test("followWallet clears the session on disconnect and never reconnects", () => {
	writeSession(alice);
	const { wallet, fire } = eventfulWallet();
	let connects = 0;
	wallet.provider.connect = async () => {
		connects++;
		return { status: "Rejected" };
	};
	const stop = follow(wallet);
	fire("disconnect", null);
	stop();
	assert.equal(readSession(), null);
	assert.equal(connects, 0);
});

test("followWallet ignores events that change nothing or are malformed", () => {
	writeSession(alice);
	const { wallet, fire } = eventfulWallet();
	const seen: unknown[] = [];
	const off = onSessionChange((s) => seen.push(s));
	const stop = follow(wallet);
	fire("accountChanged", {
		address: alice.address,
		chainid: "dev",
		pubkey: null,
	});
	fire("accountChanged", { address: "", chainid: "dev", pubkey: null });
	fire("networkChanged", { chainid: "dev", rpc: "" });
	stop();
	off();
	assert.deepEqual(seen, []);
});

test("followWallet subscribes once the remembered wallet announces, and stops with the session", () => {
	writeSession(alice);
	const { wallet, count } = eventfulWallet();
	const announced: { wallet?: GnoWallet } = {};
	let announce = () => {};
	const stop = followWallet(
		() => announced.wallet,
		(fn) => {
			announce = fn;
			return () => {};
		},
	);
	assert.equal(count("disconnect"), 0); // not announced yet
	announced.wallet = wallet;
	announce();
	assert.equal(count("disconnect"), 1);
	announce(); // a repeat announcement does not subscribe twice
	assert.equal(count("disconnect"), 1);
	clearSession();
	assert.equal(count("disconnect"), 0);
	stop();
});

test("followWallet leaves a session from another wallet alone", () => {
	writeSession({ ...alice, rdns: "other.wallet" });
	const { wallet, count } = eventfulWallet();
	const stop = followWallet(
		(rdns) => (rdns === alice.rdns ? wallet : undefined),
		() => () => {},
	);
	assert.equal(count("accountChanged"), 0);
	stop();
});
