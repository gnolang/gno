import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { test } from "node:test";
import { heredocDelimiter } from "./shell.ts";

test("heredocDelimiter never equals a line of the code", () => {
	const cases: [code: string, want: string][] = [
		["package main\n", "GNO_EOF"],
		["GNO_EOF\n", "GNO_EOF_1"],
		["a\nGNO_EOF\nGNO_EOF_1\nb", "GNO_EOF_2"],
		["  GNO_EOF\n// GNO_EOF", "GNO_EOF"],
	];
	for (const [code, want] of cases) {
		assert.equal(heredocDelimiter(code), want, code);
	}
});

test("a heredoc with heredocDelimiter writes the code whole and runs none of it", () => {
	const code = "package main\n__GNO_EOF__\nGNO_EOF\necho PWNED\n";
	const delim = heredocDelimiter(code);
	const out = execFileSync(
		"sh",
		["-c", `cat <<'${delim}'\n${code}${delim}\n`],
		{ encoding: "utf8" },
	);
	assert.equal(out, code);
});
