import "./dom.js";
import assert from "node:assert/strict";
import { test } from "node:test";
import { chooserErrorView } from "../js/wallet-chooser.js";

test("a known code gets a plain sentence and keeps the code", () => {
	assert.deepEqual(chooserErrorView("Adena", { code: "no_signer" }), {
		title: "Adena couldn't sign",
		message:
			"Adena has no account to sign with. Create or import one in Adena, then click Execute again.",
		code: "no_signer",
	});
});

test("network_declined names the page's chain when there is one", () => {
	const view = chooserErrorView("Adena", { code: "network_declined" }, "dev");
	assert.ok(view.message.includes("(dev)"));
	assert.ok(
		!chooserErrorView("Adena", { code: "network_declined" }).message.includes(
			"()",
		),
	);
});

test("an unknown or missing code gets the generic sentence and no code", () => {
	for (const err of [
		{ code: "<img src=x>" },
		new Error("boom"),
		null,
		{ code: 42 },
	]) {
		const view = chooserErrorView("Adena", err);
		assert.equal(view.message, "Adena couldn't handle this request.");
		assert.equal(view.code, null);
	}
});
