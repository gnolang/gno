import "./dom.js";
import assert from "node:assert/strict";
import { test } from "node:test";
import { helpFunc, outcomeMessage, withoutOutcome } from "../js/tx-outcome.js";

const addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5";
const base = "http://127.0.0.1:8888/r/demo/bank$help&func=Transfer&to=g1abc";

test("helpFunc reads func from the $help web query, or the query string", () => {
	assert.equal(helpFunc(new URL(base)), "Transfer");
	assert.equal(
		helpFunc(new URL("http://x/fixture.html?func=Transfer")),
		"Transfer",
	);
	assert.equal(helpFunc(new URL("http://x/r/demo/bank")), "");
});

test("signer_unavailable for this function yields a notice naming the account", () => {
	const msg = outcomeMessage(
		`${base}?status=error&code=signer_unavailable`,
		"Transfer",
		addr,
	);
	assert.ok(msg?.includes("g1jg8mtu…sqf5"));
});

test("without a session the notice drops the address", () => {
	const msg = outcomeMessage(
		`${base}?status=error&code=signer_unavailable`,
		"Transfer",
		null,
	);
	assert.ok(msg && !msg.includes("g1"));
});

test("other codes, other statuses and other functions yield nothing", () => {
	for (const href of [
		`${base}?status=error&code=user_rejected`,
		`${base}?status=success&hash=ABC`,
		`${base}?status=error`,
		base,
	]) {
		assert.equal(outcomeMessage(href, "Transfer", addr), null, href);
	}
	assert.equal(
		outcomeMessage(
			`${base}?status=error&code=signer_unavailable`,
			"Mint",
			addr,
		),
		null,
	);
});

test("an injected code is never echoed", () => {
	const href = `${base}?status=error&code=%3Cimg%20src%3Dx%3E`;
	assert.equal(outcomeMessage(href, "Transfer", addr), null);
});

test("withoutOutcome strips status and code only", () => {
	assert.equal(
		withoutOutcome(
			`${base}?status=error&code=signer_unavailable&keep=1#func-Transfer`,
		),
		`${base}?keep=1#func-Transfer`,
	);
});
