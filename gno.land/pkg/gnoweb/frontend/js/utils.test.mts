import assert from "node:assert/strict";
import { test } from "node:test";
import { evalArgs, gnoLiteral, zeroLiteral } from "./utils.ts";

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
			args: ['"a"', '"b"'],
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
