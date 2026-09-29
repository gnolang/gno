import "./dom.js";
import assert from "node:assert/strict";
import { beforeEach, test } from "node:test";
import {
	clearSession,
	type GnoSession,
	onSessionChange,
	SESSION_EVENT,
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
