import "./dom.js";
import assert from "node:assert/strict";
import { beforeEach, test } from "node:test";
import {
	abciQuery,
	ChainError,
	normalizeRPC,
	qevalJSON,
	rpcURL,
} from "../js/chain.js";
import { abciBody, installDOM, stubFetch } from "./dom.js";

beforeEach(() => {
	installDOM({ "gnoconnect:rpc": "127.0.0.1:26657" });
});

test("normalizeRPC adds a scheme and drops trailing slashes", () => {
	assert.equal(normalizeRPC("127.0.0.1:26657"), "http://127.0.0.1:26657");
	assert.equal(normalizeRPC("https://rpc.gno.land/"), "https://rpc.gno.land");
	assert.equal(normalizeRPC("  "), null);
});

test("rpcURL reads the page meta", () => {
	assert.equal(rpcURL(), "http://127.0.0.1:26657");
	installDOM({});
	assert.equal(rpcURL(), null);
});

test("abciQuery sends the path encoded and data as hex", async () => {
	const calls = stubFetch(() => ({ body: abciBody("ok") }));
	assert.equal(await abciQuery("vm/qeval", "a+b"), "ok");
	assert.equal(
		calls[0],
		"http://127.0.0.1:26657/abci_query?path=vm%2Fqeval&data=0x612b62",
	);
});

test("abciQuery returns null when the node sends no data", async () => {
	stubFetch(() => ({ body: abciBody(null) }));
	assert.equal(await abciQuery("auth/accounts/g1x"), null);
});

test("abciQuery surfaces a node error with its value", async () => {
	stubFetch(() => ({ body: abciBody(null, "unknown realm") }));
	await assert.rejects(abciQuery("vm/qeval", "x"), (err: unknown) => {
		assert.ok(err instanceof ChainError);
		assert.equal(err.nodeError, "unknown realm");
		return true;
	});
});

test("abciQuery turns HTTP and network failures into ChainError", async () => {
	stubFetch(() => ({ status: 502, body: {} }));
	await assert.rejects(abciQuery("x"), (err: unknown) => {
		assert.ok(err instanceof ChainError);
		assert.equal(err.nodeError, undefined);
		return true;
	});
	stubFetch(() => new TypeError("Failed to fetch"));
	await assert.rejects(abciQuery("x"), ChainError);
});

test("abciQuery refuses to run without an RPC endpoint", async () => {
	installDOM({});
	const calls = stubFetch(() => ({ body: abciBody("ok") }));
	await assert.rejects(abciQuery("x"), ChainError);
	assert.equal(calls.length, 0);
});

test("qevalJSON returns the first result's value", async () => {
	const result = JSON.stringify({
		results: [{ T: { "@type": "/gno.PrimitiveType" }, V: { value: "moul" } }],
	});
	stubFetch(() => ({ body: abciBody(result) }));
	assert.equal(await qevalJSON('pkg.F("x")'), "moul");
});

test("qevalJSON rejects a result that is not a string value", async () => {
	stubFetch(() => ({
		body: abciBody('{"results":[{"V":{"ObjectID":"1"}}]}'),
	}));
	await assert.rejects(qevalJSON("pkg.F()"), ChainError);
});
