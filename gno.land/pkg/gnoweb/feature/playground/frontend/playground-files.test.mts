import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
	mkdtempSync,
	readFileSync,
	rmSync,
	statSync,
	writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
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
		["$(id).gno", false],
		["a$b.gno", false],
		["`id`.gno", false],
		["a;b.gno", false],
		["a b.gno", false],
		["../x.gno", false],
		["x/y.gno", false],
		[".hidden.gno", false],
		["-rf.gno", false],
		["é.gno", false],
		["README.md", false],
		["main.gnoo", false],
		["a.gno.sh", false],
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
		['("a" string) and more', '("a" string) and more'],
		['more ("a" string)', 'more ("a" string)'],
		['("a\\x00b" string)', "a\u0000b"],
		['("\\q" string)', '("\\q" string)'],
		['("\\U00110000" string)', '("\\U00110000" string)'],
		// A result past the server's cap keeps its quotes (capEvalResult).
		[
			'("a\\n… (truncated: first 1 of 2 bytes shown)" string)',
			"a\n… (truncated: first 1 of 2 bytes shown)",
		],
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

// Each header is a valid ustar header field by field, so strict readers
// (GNU tar, Go's archive/tar) take it too, not only lenient ones.
test("createTar writes ustar headers field by field", (t) => {
	t.mock.method(Date, "now", () => 1_700_000_000_000);
	const content = "package main\n";
	const tar = createTar([{ name: "main.gno", content }]);
	const field = (at: number, len: number) =>
		new TextDecoder().decode(tar.subarray(at, at + len));
	assert.equal(field(100, 8), "0000644\0"); // mode
	assert.equal(field(108, 8), "0000000\0"); // uid
	assert.equal(field(116, 8), "0000000\0"); // gid
	assert.equal(
		field(124, 12),
		`${content.length.toString(8).padStart(11, "0")}\0`,
	); // size
	assert.equal(
		field(136, 12),
		`${(1_700_000_000).toString(8).padStart(11, "0")}\0`,
	); // mtime
	assert.equal(field(156, 1), "0"); // regular file
	assert.equal(field(257, 6), "ustar\0"); // magic
	assert.equal(field(263, 2), "00"); // version
	// The checksum is the sum of the header bytes, its own 8 read as spaces.
	let sum = 0;
	for (let i = 0; i < 512; i++) sum += i >= 148 && i < 156 ? 0x20 : tar[i];
	assert.equal(field(148, 8), `${sum.toString(8).padStart(6, "0")}\0 `);
});

test("createTar leaves out a file it cannot name, and says so", (t) => {
	const logged = t.mock.method(console, "error", () => {});
	const tar = createTar([{ name: "$(id).gno", content: "x" }]);
	assert.equal(tar.length, 512 * 2); // only the two closing blocks
	assert.match(String(logged.mock.calls[0].arguments[0]), /\$\(id\)\.gno/);
});

// The archive Download builds is one a real tar reads: names, contents,
// mode and size come back, and a corrupt header would be refused.
test("createTar writes an archive tar extracts as written", () => {
	const files = [
		{ name: "main.gno", content: "package main\n\nfunc main() {}\n" },
		{ name: "gnomod.toml", content: 'module = "gno.land/r/x"\n' },
		{
			name: "é_test.gno".replace("é", "e"),
			content: "package main // ünïcode\n",
		},
	];
	const dir = mkdtempSync(join(tmpdir(), "gnoweb-tar-"));
	try {
		writeFileSync(join(dir, "a.tar"), createTar(files));
		const listing = execFileSync("tar", ["-tvf", "a.tar"], {
			cwd: dir,
		}).toString();
		execFileSync("tar", ["-xf", "a.tar"], { cwd: dir });
		for (const f of files) {
			assert.equal(readFileSync(join(dir, f.name), "utf8"), f.content, f.name);
			assert.equal(statSync(join(dir, f.name)).mode & 0o777, 0o644, f.name);
			assert.match(listing, new RegExp(`${f.name.replace(".", "\\.")}$`, "m"));
		}
	} finally {
		rmSync(dir, { recursive: true, force: true });
	}
});

// The server opens a link of 8,000 characters at most, whose code inflates
// to 64 KiB at most: Share refuses any link it would not open, over either.
test("shareRefusal refuses the links the server would not open", () => {
	assert.equal(shareRefusal(MAX_SHARE_CODE_BYTES, MAX_SHARE_URL_LENGTH), "");
	assert.match(
		shareRefusal(MAX_SHARE_CODE_BYTES + 1, 100),
		/a link opens 64 KiB of code at most, this is 65 KiB\./,
	);
	assert.match(
		shareRefusal(100, MAX_SHARE_URL_LENGTH + 1),
		/too large to share via URL/,
	);
	for (const refusal of [
		shareRefusal(MAX_SHARE_CODE_BYTES + 1, 100),
		shareRefusal(100, MAX_SHARE_URL_LENGTH + 1),
	]) {
		assert.match(refusal, /splitting into a deployed package\.$/);
	}
});
