export function debounce<P extends unknown[]>(
	func: (...args: P) => void,
	delay = 250,
): (...args: P) => void {
	let timeoutId: ReturnType<typeof setTimeout> | undefined;
	return function (this: unknown, ...args: P) {
		if (timeoutId !== undefined) clearTimeout(timeoutId);
		timeoutId = setTimeout(() => func.apply(this, args), delay);
	};
}

// Escape for $'...' bash quoting (handles newlines + all shell special chars)
export function escapeShellSpecialChars(arg: string): string {
	return arg
		.replace(/([$`"\\!|&;<>*?{}()'])/g, "\\$1") // escape shell special chars + single quote
		.replace(/\n/g, "\\n"); // convert newlines to \n
}

// shellQuote returns s as a single shell word: as is when it holds only
// characters no shell (sh, bash, zsh) treats specially, single-quoted
// otherwise, so a key name with a space or a "$(...)" stays one argument.
export function shellQuote(s: string): string {
	if (/^[\w@%+:,./-]+$/.test(s)) return s;
	return `'${s.replace(/'/g, "'\\''")}'`;
}

export function toKebabCase(str: string): string {
	return str
		.replace(/([a-z])([A-Z])/g, "$1-$2") // replace camelCase with dash-separated words
		.toLowerCase(); // convert to lowercase
}

export function toCamelCase(str: string): string {
	str = str.replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
	return str.charAt(0).toLowerCase() + str.slice(1);
}

export function findFirstInclusive(
	root: HTMLElement,
	selector: string,
): HTMLElement | null {
	return root.matches(selector) ? root : root.querySelector(selector);
}

export function findAllInclusive(
	root: HTMLElement,
	selector: string,
): HTMLElement[] {
	const result: HTMLElement[] = [];
	if (root.matches(selector)) result.push(root);
	result.push(
		...(Array.from(root.querySelectorAll(selector)) as HTMLElement[]),
	);
	return result;
}

// EvalResult is one /_/api/eval answer. isError marks the chain's answer to
// the expression (it failed); unavailable marks a request gnoweb or the node
// could not answer (rate limit, outage, network, abort), which says nothing
// about the expression and is shown without the "Error:" prefix.
export interface EvalResult {
	text: string;
	isError: boolean;
	unavailable: boolean;
}

// postEval evaluates expression in pkgPath through gnoweb's /_/api/eval.
// signal aborts a request a newer one supersedes.
export async function postEval(
	pkgPath: string,
	expression: string,
	signal?: AbortSignal,
): Promise<EvalResult> {
	try {
		const response = await fetch("/_/api/eval", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ pkg_path: pkgPath, expression }),
			signal,
		});
		const json = await response.json().catch(() => ({}));
		const { status } = response;
		if (status === 429 || status >= 500) {
			const text = json.error || `HTTP ${status}`;
			return { text, isError: false, unavailable: true };
		}
		if (json.error) {
			const text = `Error: ${json.error}`;
			return { text, isError: true, unavailable: false };
		}
		if (!response.ok) {
			const text = `Error: HTTP ${status}`;
			return { text, isError: true, unavailable: false };
		}
		const text = json.result ?? "";
		return { text, isError: false, unavailable: false };
	} catch {
		const text = "gnoweb could not be reached, try again shortly";
		return { text, isError: false, unavailable: true };
	}
}

// gnoStringLiteral matches exactly one Gno string literal: interpreted,
// with only the escapes Go's spec allows, or raw.
const gnoStringLiteral =
	/^(?:"(?:[^"\\\n]|\\(?:[abfnrtv\\"]|[0-7]{3}|x[0-9A-Fa-f]{2}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8}))*"|`[^`]*`)$/;

// gnoOperand matches one operand that is no expression: a signed number,
// a rune literal, or a name such as true, nil or a constant.
const gnoOperand =
	/^(?:[+-]?\.?[0-9][0-9A-Za-z_.]*(?:[eEpP][+-][0-9_]+)?|'(?:[^'\\\n]|\\[^\n][0-9A-Fa-f]{0,7})'|[A-Za-z_][A-Za-z0-9_]*)$/;

// gnoLiteral turns a typed value into one Gno operand for its param type,
// never an expression: a link's prefilled value must not make a card
// evaluate (and show as its function's result) anything but that function.
// A string or an address is kept only when it is exactly one string
// literal; any other type also keeps a number, a rune or a name. Anything
// else is quoted, so it reaches the function as a string.
export function gnoLiteral(value: string, type: string): string {
	const asIs =
		gnoStringLiteral.test(value) ||
		(type !== "string" && type !== "address" && gnoOperand.test(value));
	return asIs ? value : JSON.stringify(value);
}

// zeroLiteral is a Gno literal of the zero value for a parameter type, as
// the package's doc renders it: "" for strings and addresses, 0 for
// numbers, false for bool and nil for the rest, a placeholder the reader
// replaces.
export function zeroLiteral(type: string): string {
	if (type === "string" || type === "address") return '""';
	if (type === "bool") return "false";
	if (/^(u?int(8|16|32|64)?|uintptr|byte|rune|float(32|64))$/.test(type))
		return "0";
	return "nil";
}

// evalArgs turns a function card's param inputs, in order, into Gno
// arguments: each typed value as a literal for the input's param type, each
// empty one as its zero value. The template renders one input per param,
// and every one counts, so unnamed or repeated names ("_") keep their
// argument. full is false while any param is empty.
export function evalArgs(
	inputs: Pick<HTMLInputElement, "value" | "getAttribute">[],
): { args: string[]; full: boolean } {
	let full = true;
	const args = inputs.map((input) => {
		const type = input.getAttribute("data-action-function-type-value") ?? "";
		const value = input.value.trim();
		if (!value) full = false;
		return value ? gnoLiteral(value, type) : zeroLiteral(type);
	});
	return { args, full };
}

// downloadBlob saves blob as filename through a temporary object URL.
export function downloadBlob(blob: Blob, filename: string): void {
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = filename;
	a.click();
	URL.revokeObjectURL(url);
}

export function makeIcon(icon: string): SVGSVGElement {
	const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
	svg.classList.add("c-icon");
	const use = document.createElementNS("http://www.w3.org/2000/svg", "use");
	use.setAttribute("href", `#ico-${icon}`);
	svg.appendChild(use);
	return svg;
}

export function makeCopyIcon(): SVGSVGElement {
	const svg = makeIcon("copy");
	svg.querySelector("use")?.setAttribute("data-copy-target", "icon");

	const checkUse = document.createElementNS(
		"http://www.w3.org/2000/svg",
		"use",
	);
	checkUse.setAttribute("href", "#ico-check");
	checkUse.setAttribute("data-copy-target", "icon");
	checkUse.classList.add("u-hidden", "u-color-valid");
	svg.appendChild(checkUse);
	return svg;
}

// setPrefCookie writes a SameSite=Lax preference cookie, adding Secure
// on HTTPS. Used by controllers that mirror a JS preference to a cookie
// so the server can stamp the right state on first paint.
export function setPrefCookie(
	name: string,
	value: string,
	maxAge: number,
): void {
	const secure = location.protocol === "https:" ? ";Secure" : "";
	document.cookie = `${name}=${value};path=/;max-age=${maxAge};SameSite=Lax${secure}`;
}
