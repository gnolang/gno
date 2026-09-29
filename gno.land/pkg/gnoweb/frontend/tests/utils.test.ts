import "./dom.js";
import assert from "node:assert/strict";
import { test } from "node:test";
import { toCamelCase, toKebabCase } from "../js/utils.js";

test("toKebabCase splits camelCase", () => {
	assert.equal(toKebabCase("actionFunction"), "action-function");
});

test("toCamelCase joins kebab-case", () => {
	assert.equal(toCamelCase("wallet-launch"), "walletLaunch");
});
