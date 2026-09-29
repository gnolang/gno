import "./dom.js";
import assert from "node:assert/strict";
import { beforeEach, test } from "node:test";
import {
	clearSession,
	displayName,
	type GnoSession,
	onSessionChange,
	readSession,
	SESSION_EVENT,
	setUsername,
	writeSession,
} from "../../feature/connect/frontend/session.js";
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
