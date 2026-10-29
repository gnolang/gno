// Shell-safe pieces of the command the run page prints.

// shellQuote returns s as a single shell word: as is when it holds only
// characters no shell (sh, bash, zsh) treats specially, single-quoted
// otherwise, so a key name with a space or a "$(...)" stays one argument.
export function shellQuote(s: string): string {
	if (/^[\w@%+:,./-]+$/.test(s)) return s;
	return `'${s.replace(/'/g, "'\\''")}'`;
}

// heredocDelimiter returns a heredoc delimiter that no line of code equals,
// so the heredoc carrying code ends where the code does and no line of it
// runs as a command.
export function heredocDelimiter(code: string): string {
	const lines = new Set(code.split("\n"));
	let delim = "GNO_EOF";
	for (let i = 1; lines.has(delim); i++) delim = `GNO_EOF_${i}`;
	return delim;
}
