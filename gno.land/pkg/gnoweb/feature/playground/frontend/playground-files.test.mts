import assert from "node:assert/strict";
import { test } from "node:test";
import { inflateRawSync } from "node:zlib";
import {
	copyFiles,
	createTar,
	isValidFileName,
	MAX_SHARE_CODE_BYTES,
	MAX_SHARE_URL_LENGTH,
	renderText,
	shareLink,
	shareRefusal,
} from "./playground-files.ts";

test("isValidFileName takes Gno sources and gnomod.toml from a strict allowlist", () => {
	const cases: [name: string, want: boolean][] = [
		["main.gno", true],
		["gnomod.toml", true],
		["a_test.gno", true],
		["Z-1.v2.gno", true],
		[`${"x".repeat(96)}.gno`, true], // 100 bytes
		[`${"x".repeat(97)}.gno`, false], // 101 bytes, cut by a ustar header
		["$(touch PWNED).gno", false],
		["a;b.gno", false],
		["a b.gno", false],
		["../x.gno", false],
		["x/y.gno", false],
		[".hidden.gno", false],
		["-rf.gno", false],
		["é.gno", false],
		["README.md", false],
		["Base set", false],
		[".gno", false],
		["", false],
	];
	for (const [name, want] of cases) {
		assert.equal(isValidFileName(name), want, name);
	}
});

// shareFiles reads a link back as the server does (handler.go's
// sharedFiles): base64, raw DEFLATE, then the JSON list of files.
function shareFiles(url: string): { name: string; content: string }[] {
	const u = new URL(url);
	assert.equal(u.pathname, "/_/play");
	assert.ok(u.searchParams.has("z"));
	const code = u.searchParams.get("code") ?? "";
	return JSON.parse(inflateRawSync(Buffer.from(code, "base64")).toString());
}

test("a share link reads back to the files it was written from", async () => {
	const many = Array.from({ length: 60 }, (_, i) => ({
		name: `f${i}.gno`,
		content: `package f\n\nvar N${i} = ${i}\n`,
	}));
	const cases: { name: string; content: string }[][] = [
		[{ name: "counter.gno", content: "package counter\n" }],
		[
			{
				name: "gnomod.toml",
				content: 'module = "gno.land/r/x"\ngno = "0.9"\n',
			},
		],
		// Lines shaped like the separator the text format had: no line of a
		// file is read as anything.
		[
			{
				name: "a.gno",
				content:
					"package a\n\n// --- b.gno ---\nfunc A() {}\n// --- Base set ---\n",
			},
			{ name: "b.gno", content: "// --- a.gno ---\npackage a\n" },
		],
		// Blank lines at either edge, no final newline, an empty file.
		[
			{ name: "c.gno", content: "\n\n\npackage c\n\n\n" },
			{ name: "d.gno", content: "package d" },
			{ name: "e.gno", content: "" },
		],
		[{ name: "crlf.gno", content: "package crlf\r\n\r\nfunc F() {}\r\n" }],
		[{ name: "u.gno", content: 'package u // héllo 世界 🎉 "\\" \u2028\n' }],
		many,
	];
	for (const files of cases) {
		const { url, bytes } = await shareLink("https://gno.land", files);
		assert.ok(url.startsWith("https://gno.land/_/play?code="), url);
		assert.deepEqual(shareFiles(url), files);
		assert.equal(bytes, Buffer.byteLength(JSON.stringify(files)));
		assert.equal(shareRefusal(bytes, url.length), "");
	}
});

test("a share link holds only names and contents", async () => {
	const files = [{ name: "a.gno", content: "package a\n", extra: 1 }];
	const { url } = await shareLink("", files);
	assert.deepEqual(shareFiles(`https://x${url}`), [
		{ name: "a.gno", content: "package a\n" },
	]);
});

test("Share refuses a link over either cap", async () => {
	// Compressible past 64 KiB of JSON, within 8,000 characters of link.
	const big = await shareLink("", [
		{ name: "a.gno", content: "a".repeat(MAX_SHARE_CODE_BYTES) },
	]);
	assert.ok(big.bytes > MAX_SHARE_CODE_BYTES);
	assert.ok(big.url.length <= MAX_SHARE_URL_LENGTH);
	assert.match(
		shareRefusal(big.bytes, big.url.length),
		/64 KiB of code at most/,
	);
	// Incompressible: under 64 KiB, but the link is too long.
	let x = 1;
	const noise = Array.from({ length: 20_000 }, () => {
		x = (x * 1103515245 + 12345) % 2 ** 31;
		return String.fromCharCode(33 + (x % 90));
	}).join("");
	const long = await shareLink("", [{ name: "a.gno", content: noise }]);
	assert.ok(long.bytes <= MAX_SHARE_CODE_BYTES);
	assert.match(
		shareRefusal(long.bytes, long.url.length),
		/too large to share via URL/,
	);
});

test("copyFiles keeps the page's files one by one", () => {
	const files = [
		{ name: "a.gno", content: "package a\n\n// --- b.gno ---\nfunc A() {}\n" },
		{ name: "b.gno", content: "package a\n" },
	];
	assert.deepEqual(copyFiles(files), { files, skipped: [] });
	// Names the chain takes but the playground does not are reported.
	assert.deepEqual(
		copyFiles([
			{ name: "../x.gno", content: "" },
			{ name: "-x.gno", content: "" },
			{ name: "b.gno", content: "" },
			{ name: "b.gno", content: "again" },
		]),
		{
			files: [{ name: "b.gno", content: "" }],
			skipped: ["../x.gno", "-x.gno", "b.gno"],
		},
	);
});

test("renderText reads the string a Render eval returns", () => {
	const cases: [result: string, want: string][] = [
		['("0" string)', "0"],
		['("# Title\\n\\nsay \\"hi\\"" string)', '# Title\n\nsay "hi"'],
		[
			'("tab\\there\\x1b\\v\\a\\x7f end" string)',
			"tab\there\x1b\v\x07\x7f end",
		],
		['("\\U0001f600 \\xff \\\\x1b" string)', "\u{1f600} \ufffd \\x1b"],
		["(42 int)", "(42 int)"],
		["Error: package not found", "Error: package not found"],
	];
	for (const [result, want] of cases) {
		assert.equal(renderText(result), want, result);
	}
});

test("createTar writes whole names and skips the ones it cannot hold", () => {
	const name = `${"x".repeat(96)}.gno`;
	const tar = createTar([
		{ name, content: "package x\n" },
		{ name: `${"y".repeat(97)}.gno`, content: "package y\n" },
	]);
	// One header and one data block, then the two closing blocks.
	assert.equal(tar.length, 512 * 4);
	assert.equal(new TextDecoder().decode(tar.subarray(0, 100)), name);
});

// The server opens a link of 8,000 characters at most, whose code inflates
// to 64 KiB at most: Share refuses any link it would not open, over either.
test("shareRefusal refuses the links the server would not open", () => {
	assert.equal(shareRefusal(MAX_SHARE_CODE_BYTES, MAX_SHARE_URL_LENGTH), "");
	assert.match(
		shareRefusal(MAX_SHARE_CODE_BYTES + 1, 100),
		/64 KiB of code at most, this is 65 KiB/,
	);
	assert.match(
		shareRefusal(100, MAX_SHARE_URL_LENGTH + 1),
		/too large to share via URL/,
	);
});
