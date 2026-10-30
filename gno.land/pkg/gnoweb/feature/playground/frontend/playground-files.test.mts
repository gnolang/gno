import assert from "node:assert/strict";
import { test } from "node:test";
import {
	copyFiles,
	createTar,
	isValidFileName,
	MAX_SHARE_CODE_BYTES,
	MAX_SHARE_URL_LENGTH,
	parseInitialCode,
	renderText,
	shareRefusal,
	shareText,
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

test("parseInitialCode keeps a comment shaped like a header in its file", () => {
	// p/nt/groups/v0's group.gno holds such section comments.
	const group =
		"package groups\n\n// --- Base set ---\n\nfunc A() {}\n\n// --- Role registry ---\nfunc B() {}\n";
	const files = parseInitialCode(
		`// --- group.gno ---\n${group}\n\n// --- role.gno ---\npackage groups\n`,
	);
	assert.deepEqual(files, [
		{ name: "group.gno", content: group },
		{ name: "role.gno", content: "package groups\n" },
	]);
});

test("parseInitialCode reads text before the first header as main.gno", () => {
	assert.deepEqual(
		parseInitialCode(
			"package main\n\nfunc main() {}\n// --- foo.gno ---\npackage main\n",
		),
		[
			{ name: "main.gno", content: "package main\n\nfunc main() {}\n" },
			{ name: "foo.gno", content: "package main\n" },
		],
	);
	// A rejected header with nothing before it still keeps its text.
	assert.deepEqual(parseInitialCode("// --- not a file ---\npackage main\n"), [
		{ name: "main.gno", content: "// --- not a file ---\npackage main\n" },
	]);
});

test("parseInitialCode keeps a repeated name as text in the file before", () => {
	assert.deepEqual(
		parseInitialCode(
			"// --- a.gno ---\npackage a\n// --- a.gno ---\nfunc A() {}\n",
		),
		[{ name: "a.gno", content: "package a\n// --- a.gno ---\nfunc A() {}\n" }],
	);
});

test("parseInitialCode leaves code without headers as it is", () => {
	assert.deepEqual(parseInitialCode("package main"), [
		{ name: "main.gno", content: "package main" },
	]);
});

test("a share reads back to the files it was written from, a lone file too", () => {
	const cases: { name: string; content: string }[][] = [
		[
			{
				name: "gnomod.toml",
				content: 'module = "gno.land/r/x"\ngno = "0.9"\n',
			},
		],
		[{ name: "counter.gno", content: "package counter\n" }],
		[
			{
				name: "main.gno",
				content: "package main\n\n// --- note ---\nfunc main() {}\n",
			},
			{ name: "gnomod.toml", content: 'module = "main"\n' },
		],
	];
	for (const files of cases) {
		assert.deepEqual(parseInitialCode(shareText(files)), files);
	}
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
