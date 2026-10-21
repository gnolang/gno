import "./dom.js";
import assert from "node:assert/strict";
import { test } from "node:test";
import { linkIsSelf } from "../js/you-badge.js";

const me = "g1manfred47kzduec920z88wfr64ylksmdcedlf5";
const base = "http://127.0.0.1:8888/r/gnoland/coins:balances";

test("a /u/ link to my username is me", () => {
	assert.ok(linkIsSelf("/u/moul", base, me, "moul"));
	assert.equal(linkIsSelf("/u/moul2", base, me, "moul"), false);
	assert.equal(linkIsSelf("/u/moul", base, me), false);
});

test("a link carrying my address is me, wherever it points", () => {
	assert.ok(linkIsSelf(`/r/demo/defi/foo20:balance/${me}`, base, me));
	assert.ok(linkIsSelf(`https://explorer.example/account/${me}`, base, me));
	assert.equal(linkIsSelf("/r/demo/defi/foo20", base, me), false);
});

test("an unparseable href is not me", () => {
	assert.equal(linkIsSelf("http://[::1", base, me, "moul"), false);
});

test("a link carrying my address only in its query string is not me", () => {
	// Table sort links keep the page's ?address= state; they point at a view.
	assert.equal(
		linkIsSelf(`/r/gnoland/coins:balances?address=${me}&sort=denom`, base, me),
		false,
	);
});

test("a malformed percent-escape is not me and does not throw", () => {
	assert.equal(linkIsSelf("/u/%E0%A4%A", base, me, "moul"), false);
});
