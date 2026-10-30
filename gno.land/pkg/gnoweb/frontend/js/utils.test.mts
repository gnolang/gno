import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { test } from "node:test";
import {
	addressSlotText,
	argSlotText,
	evalArgs,
	gnoLiteral,
	postEval,
	shellQuote,
	zeroLiteral,
} from "./utils.ts";

test("gnoLiteral quotes strings and addresses, and leaves the rest as typed", () => {
	const cases: [value: string, type: string, want: string][] = [
		["He11o", "string", '"He11o"'],
		['a "q" \\ b', "string", '"a \\"q\\" \\\\ b"'],
		["line\nbreak", "string", '"line\\nbreak"'],
		['"He11o"', "string", '"He11o"'],
		['"a \\" b"', "string", '"a \\" b"'],
		["`raw`", "string", "`raw`"],
		['"open', "string", '"\\"open"'],
		[
			"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			"address",
			'"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"',
		],
		["42", "int", "42"],
		["-7", "int64", "-7"],
		["0x1F", "uint", "0x1F"],
		["1e+5", "float64", "1e+5"],
		["'a'", "rune", "'a'"],
		["'\\n'", "rune", "'\\n'"],
		["true", "bool", "true"],
		["nil", "ProposalID", "nil"],
		["MaxSupply", "int", "MaxSupply"],
		['"x"', "Name", '"x"'],
		["raw", "", "raw"],
	];
	for (const [value, type, want] of cases) {
		assert.equal(gnoLiteral(value, type), want, `${type} ${value}`);
	}
});

// A prefilled link must not make a card evaluate another expression and
// show its value as the function's result: anything that is not exactly one
// operand reaches the function as a string.
test("gnoLiteral quotes any value that is more than one operand", () => {
	const cases: [value: string, type: string][] = [
		['"a"+Other()+"b"', "string"],
		['"a" + Other() + "b"', "address"],
		["`a`+Other()+`b`", "string"],
		['"a"//x', "string"],
		['"\\q"', "string"], // not an escape Gno has
		['"\\x4"', "string"],
		["Other()", "string"],
		["Other()", "int"],
		["1+Other()", "int"],
		["-Other()", "int"],
		["pkg.Secret", "int"],
		["'a'+Other()", "rune"],
		["NewThing()", "ProposalID"],
		["[]string{Other()}", "[]string"],
		["true || Other()", "bool"],
	];
	for (const [value, type] of cases) {
		assert.equal(
			gnoLiteral(value, type),
			JSON.stringify(value),
			`${type} ${value}`,
		);
	}
	// One literal holding what looks like a call is still that one literal.
	assert.equal(
		gnoLiteral('"a\\"+Other()+\\"b"', "string"),
		'"a\\"+Other()+\\"b"',
	);
	assert.equal(
		gnoLiteral('"\\x41\\u00e9\\101"', "string"),
		'"\\x41\\u00e9\\101"',
	);
});

test("zeroLiteral gives the zero value the doc renders for a type", () => {
	const cases: [type: string, want: string][] = [
		["string", '""'],
		["address", '""'],
		["bool", "false"],
		["int", "0"],
		["uint64", "0"],
		["float64", "0"],
		["rune", "0"],
		["ProposalID", "nil"],
		["[]string", "nil"],
	];
	for (const [type, want] of cases) {
		assert.equal(zeroLiteral(type), want, type);
	}
});

// input fakes a param input as action.html renders it.
function input(name: string, type: string, value: string) {
	const attrs: Record<string, string> = {
		"data-action-function-param-value": name,
		"data-action-function-type-value": type,
	};
	return { value, getAttribute: (attr: string) => attrs[attr] ?? null };
}

test("evalArgs gives every param its argument, in order", () => {
	const cases: {
		name: string;
		inputs: ReturnType<typeof input>[];
		args: string[];
		full: boolean;
	}[] = [
		{ name: "no params", inputs: [], args: [], full: true },
		{
			name: "unnamed param",
			inputs: [input("", "string", "")],
			args: ['""'],
			full: false,
		},
		{
			name: "repeated names, all empty",
			inputs: [
				input("_", "int", ""),
				input("_", "ProposalID", ""),
				input("_", "[]string", ""),
				input("", "string", ""),
			],
			args: ["0", "nil", "nil", '""'],
			full: false,
		},
		{
			name: "repeated names, all typed",
			inputs: [input("_", "string", "a"), input("_", "string", " b ")],
			args: ['"a"', '" b "'],
			full: true,
		},
		// The printed commands carry a value as typed: so do the live
		// result and Eval, and only an empty input is empty.
		{
			name: "spaces are kept",
			inputs: [input("s", "string", "  "), input("n", "int", " 5")],
			args: ['"  "', '" 5"'],
			full: true,
		},
		{
			name: "typed string and address",
			inputs: [
				input("name", "string", "He11o"),
				input("addr", "address", "g1abc"),
				input("n", "int", "3"),
			],
			args: ['"He11o"', '"g1abc"', "3"],
			full: true,
		},
	];
	for (const c of cases) {
		assert.deepEqual(
			evalArgs(c.inputs),
			{ args: c.args, full: c.full },
			c.name,
		);
	}
});

// bashArgs returns the arguments bash passes to a command line's first
// word, NUL-separated so an argument may hold a newline. History expansion
// is on, as in the interactive shell these commands are pasted into.
function bashArgs(line: string): string[] {
	const out = execFileSync(
		"bash",
		["-c", `set -o history -H\ncmd() { printf '%s\\0' "$@"; }; ${line}`],
		{ encoding: "utf8" },
	);
	return out.split("\0").slice(0, -1);
}

test("shellQuote leaves plain values bare", () => {
	for (const s of ["test1", "50_000_000", "100000ugnot", "1ugnot,2foo"]) {
		assert.equal(shellQuote(s), s);
	}
});

const awkwardKeys = [
	"my key",
	'1ugnot"$(id)"',
	"$(id)",
	"$HOME",
	"it's",
	"`id`",
	"a;b|c&d",
	"~root",
	"#x",
	"=ls",
	"*",
	"",
];

test("shellQuote keeps any value one literal argument", () => {
	for (const s of awkwardKeys) {
		assert.deepEqual(bashArgs(`cmd ${shellQuote(s)}`), [s], s);
	}
});

// $help prints the key or address as the last word of its gnokey lines, and
// after auth/accounts/: each must reach gnokey whole, as typed.
test("the key in $help's commands reaches gnokey as typed", () => {
	for (const key of awkwardKeys.filter(Boolean)) {
		const k = shellQuote(key);
		assert.deepEqual(
			bashArgs(`cmd sign -tx-path call.tx ${k}`),
			["sign", "-tx-path", "call.tx", key],
			key,
		);
		assert.deepEqual(
			bashArgs(`cmd query auth/accounts/${k}`),
			["query", `auth/accounts/${key}`],
			key,
		);
	}
});

// A maketx call arg slot has no type: gnokey gets the value as typed, never
// a Gno literal of it. A qeval slot has one, and gets the literal.
test("$help's maketx args reach gnokey as typed", () => {
	for (const v of [
		"bob smith",
		"hello world-1",
		"my-name",
		"a,b",
		"https://x.y/z",
		"it's",
		'say "hi"',
		"alice",
		"1000ugnot",
	]) {
		assert.deepEqual(bashArgs(`cmd $'${argSlotText(v, "")}'`), [v], v);
	}
	assert.equal(argSlotText("", ""), "");
});

// Values a reader may type that a shell would read as syntax: each must
// reach gnokey as the literal (qeval) or the value (maketx) it stands for.
const hostileArgs = [
	"a(b) hi!",
	"!",
	"!!",
	"a!b",
	"$(id)",
	"`id`",
	"$" + "{HOME}", // split so it reads as the text it is
	"\\",
	'"',
	"'",
	"\\n",
	"{a,b}",
	"* ? | & ; < > ( )",
	'say "hi" it\'s $x\\',
	"`raw` literal",
	"line\nbreak",
	"it's done!",
	"it's !!",
	"it's a!b",
];

// A qeval slot sits inside -data "pkg.F(...)", in double quotes, where a
// backslash before anything but \ " $ ` stays in the word: $'...' escaping
// would leave it before every ( ) ' ! and break the call.
test("$help's qeval command reaches gnokey with the Eval button's literal", () => {
	for (const type of ["string", "int", "[]string"]) {
		for (const v of hostileArgs) {
			const slot = argSlotText(v, type);
			assert.deepEqual(
				bashArgs(`cmd -data "gno.land/r/x.F(${slot})"`),
				["-data", `gno.land/r/x.F(${gnoLiteral(v, type)})`],
				`${type} ${v}`,
			);
		}
	}
});

test("$help's maketx args reach gnokey as typed, hostile or not", () => {
	for (const v of hostileArgs) {
		assert.deepEqual(bashArgs(`cmd $'${argSlotText(v, "")}'`), [v], v);
	}
});

test("$help's address slot is the quoted key, or ADDRESS", () => {
	assert.equal(addressSlotText(" "), "ADDRESS");
	for (const key of awkwardKeys.filter(Boolean)) {
		assert.deepEqual(
			bashArgs(`cmd ${addressSlotText(` ${key} `)}`),
			[key],
			key,
		);
	}
});

test("gnoLiteral keeps only a complete Gno literal or operand as is", () => {
	const cases: [value: string, type: string, want: string][] = [
		['"\\101"', "string", '"\\101"'], // octal needs three digits
		['"\\1"', "string", JSON.stringify('"\\1"')],
		['"\\u00e9"', "string", '"\\u00e9"'],
		['"\\u12"', "string", JSON.stringify('"\\u12"')],
		['"\\U0001f600"', "string", '"\\U0001f600"'],
		['"\\U0001f60"', "string", JSON.stringify('"\\U0001f60"')],
		["1e+10", "float64", "1e+10"],
		["'\\x41'", "rune", "'\\x41'"],
	];
	for (const [value, type, want] of cases) {
		assert.equal(gnoLiteral(value, type), want, value);
	}
});

test("zeroLiteral reads the whole type name", () => {
	for (const type of ["[]int", "int64Alias", "uintx", "*float64"]) {
		assert.equal(zeroLiteral(type), "nil", type);
	}
});

test("argSlotText leaves an empty slot empty, typed or not", () => {
	assert.equal(argSlotText("", "string"), "");
	assert.equal(argSlotText("", ""), "");
});

// postEval against a stubbed fetch: what it sends, and how each answer is
// shown. 429 and 5xx are gnoweb or the node being busy or down, shown as
// unavailable rather than as the function's error.
test("postEval sends the expression and sorts the answers", async (t) => {
	const calls: { url: string; init: RequestInit }[] = [];
	const answer = (status: number, body: unknown) =>
		new Response(body === undefined ? "not json" : JSON.stringify(body), {
			status,
		});
	const cases: [
		Response | Error,
		{ text: string; isError: boolean; unavailable: boolean },
	][] = [
		[
			answer(200, { result: '("x" string)' }),
			{ text: '("x" string)', isError: false, unavailable: false },
		],
		[answer(200, {}), { text: "", isError: false, unavailable: false }],
		[
			answer(200, { error: "boom" }),
			{ text: "Error: boom", isError: true, unavailable: false },
		],
		[
			answer(400, { error: "bad" }),
			{ text: "Error: bad", isError: true, unavailable: false },
		],
		[
			answer(404, undefined),
			{ text: "Error: HTTP 404", isError: true, unavailable: false },
		],
		[
			answer(429, { error: "slow down" }),
			{ text: "slow down", isError: false, unavailable: true },
		],
		[
			answer(499, undefined),
			{ text: "Error: HTTP 499", isError: true, unavailable: false },
		],
		[
			answer(500, undefined),
			{ text: "HTTP 500", isError: false, unavailable: true },
		],
		[
			answer(503, { error: "busy" }),
			{ text: "busy", isError: false, unavailable: true },
		],
		[
			new TypeError("offline"),
			{
				text: "gnoweb could not be reached, try again shortly",
				isError: false,
				unavailable: true,
			},
		],
	];
	const realFetch = globalThis.fetch;
	t.after(() => {
		globalThis.fetch = realFetch;
	});
	for (const [reply, want] of cases) {
		globalThis.fetch = (async (url: string, init: RequestInit) => {
			calls.push({ url, init });
			if (reply instanceof Error) throw reply;
			return reply;
		}) as typeof fetch;
		assert.deepEqual(await postEval("gno.land/r/x", "F(1)"), want);
	}
	const { url, init } = calls[0];
	assert.equal(url, "/_/api/eval");
	assert.equal(init.method, "POST");
	assert.deepEqual(init.headers, { "Content-Type": "application/json" });
	assert.deepEqual(JSON.parse(init.body as string), {
		pkg_path: "gno.land/r/x",
		expression: "F(1)",
	});
});
