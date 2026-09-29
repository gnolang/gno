import "./dom.js";
import assert from "node:assert/strict";
import { beforeEach, test } from "node:test";
import {
	accountView,
	chainMatches,
	commandNumbers,
	formatGnot,
	isAddress,
	loadAccount,
	loadUsername,
	parseAccount,
	refresh,
} from "../js/account.js";
import { ChainError } from "../js/chain.js";
import { abciBody, installDOM, stubFetch } from "./dom.js";

const addr = "g1manfred47kzduec920z88wfr64ylksmdcedlf5";
const accountJSON = (coins: string) =>
	JSON.stringify({
		BaseAccount: {
			address: addr,
			coins,
			account_number: "3096238",
			sequence: "368",
		},
	});

beforeEach(() => {
	installDOM({
		"gnoconnect:rpc": "127.0.0.1:26657",
		"gnoconnect:chainid": "dev",
	});
	refresh(addr);
});

test("isAddress accepts bech32 g1 only", () => {
	assert.ok(isAddress(addr));
	for (const bad of [
		"mykey",
		addr.toUpperCase(),
		`${addr}x`,
		"g1abc",
		`g1${"b".repeat(38)}`,
		`x") + evil("`,
	]) {
		assert.equal(isAddress(bad), false, bad);
	}
});

test("parseAccount reads number, sequence and the ugnot balance", () => {
	const a = parseAccount(addr, accountJSON("105537570490ugnot"));
	assert.deepEqual(a, {
		address: addr,
		accountNumber: "3096238",
		sequence: "368",
		ugnot: 105537570490n,
		otherDenoms: 0,
	});
});

test("parseAccount maps a missing account to null", () => {
	assert.equal(parseAccount(addr, "null"), null);
	assert.equal(parseAccount(addr, null), null);
});

test("junk denoms are counted, never kept", () => {
	const a = parseAccount(
		addr,
		accountJSON("5ugnot,1<script>,10foo/bar,7ugnot,,3atom"),
	);
	assert.equal(a?.ugnot, 12n);
	assert.equal(a?.otherDenoms, 2); // foo/bar and atom; <script> is not a denom
	assert.ok(!JSON.stringify(accountView(a)).includes("script"));
});

test("formatGnot is exact and trims trailing zeros", () => {
	assert.equal(formatGnot(0n), "0 GNOT");
	assert.equal(formatGnot(1_000_000n), "1 GNOT");
	assert.equal(formatGnot(12_500_000n), "12.5 GNOT");
	assert.equal(formatGnot(1n), "0.000001 GNOT");
	assert.equal(formatGnot(105537570490n), "105,537.57049 GNOT");
	assert.equal(formatGnot(2n ** 64n), "18,446,744,073,709.551616 GNOT");
});

test("accountView covers present, extra tokens, and not-on-chain", () => {
	const a = parseAccount(addr, accountJSON("12500000ugnot,1foo,2bar"));
	assert.deepEqual(accountView(a), {
		balance: "12.5 GNOT · +2 other tokens",
		line: "Account 3096238 · Sequence 368",
	});
	assert.deepEqual(accountView(null), {
		balance: "0 GNOT · not on chain yet",
		line: null,
	});
});

test("commandNumbers fills or keeps the placeholders", () => {
	assert.deepEqual(commandNumbers(null), {
		accountNumber: "ACCOUNTNUMBER",
		sequence: "SEQUENCENUMBER",
	});
	assert.deepEqual(commandNumbers(parseAccount(addr, accountJSON("1ugnot"))), {
		accountNumber: "3096238",
		sequence: "368",
	});
});

test("chainMatches blocks only a real mismatch", () => {
	assert.ok(chainMatches("dev"));
	assert.ok(chainMatches(""));
	assert.equal(chainMatches("gnoland-1"), false);
	installDOM({});
	assert.ok(chainMatches("gnoland-1"));
});

test("concurrent loads share one request; refresh forces a new one", async () => {
	const calls = stubFetch(() => ({ body: abciBody(accountJSON("1ugnot")) }));
	const [a, b] = await Promise.all([loadAccount(addr), loadAccount(addr)]);
	assert.equal(calls.length, 1);
	assert.equal(a, b);
	refresh(addr);
	await loadAccount(addr);
	assert.equal(calls.length, 2);
	assert.ok(calls[0].includes(`path=auth%2Faccounts%2F${addr}`));
});

test("a failed load is not cached", async () => {
	stubFetch(() => ({ status: 502, body: {} }));
	await assert.rejects(loadAccount(addr), ChainError);
	const calls = stubFetch(() => ({ body: abciBody(accountJSON("1ugnot")) }));
	await loadAccount(addr);
	assert.equal(calls.length, 1);
});

test("loadAccount rejects non-addresses without a request", async () => {
	const calls = stubFetch(() => ({ body: abciBody("null") }));
	for (const bad of ["mykey", addr.toUpperCase(), "g1abc"]) {
		await assert.rejects(loadAccount(bad), ChainError);
	}
	assert.equal(calls.length, 0);
});

test("loadUsername: registered, unregistered, and RPC down", async () => {
	stubFetch(() => ({
		body: abciBody(JSON.stringify({ results: [{ V: { value: "moul" } }] })),
	}));
	assert.equal(await loadUsername(addr), "moul");

	const other = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5";
	const calls = stubFetch(() => ({
		body: abciBody(null, "<error: runtime.errorString>"),
	}));
	assert.equal(await loadUsername(other), null);
	assert.ok(decodeURIComponent(calls[0]).includes("path=vm/qeval_json"));

	const third = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq";
	stubFetch(() => new TypeError("Failed to fetch"));
	await assert.rejects(loadUsername(third), ChainError);
});
