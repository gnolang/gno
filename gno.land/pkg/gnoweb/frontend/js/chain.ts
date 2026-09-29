// Client-side reads against the RPC the page advertises in
// <meta name="gnoconnect:rpc">. Callers degrade on ChainError.

export class ChainError extends Error {
	// nodeError is the node's own error value; unset for transport failures.
	constructor(
		message: string,
		readonly nodeError?: string,
	) {
		super(message);
		this.name = "ChainError";
	}
}

interface AbciEnvelope {
	result?: {
		response?: {
			ResponseBase?: {
				Error?: { value?: string } | null;
				Data?: string | null;
			};
		};
	};
}

export function readMeta(name: string): string {
	return (
		document
			.querySelector(`meta[name="${name}"]`)
			?.getAttribute("content")
			?.trim() ?? ""
	);
}

export function normalizeRPC(raw: string): string | null {
	const value = raw.trim();
	if (!value) return null;
	const url = /^[a-z][a-z0-9+.-]*:\/\//i.test(value)
		? value
		: `http://${value}`;
	return url.replace(/\/+$/, "");
}

export function rpcURL(): string | null {
	return normalizeRPC(readMeta("gnoconnect:rpc"));
}

function toHex(text: string): string {
	return Array.from(new TextEncoder().encode(text), (b) =>
		b.toString(16).padStart(2, "0"),
	).join("");
}

function fromBase64(data: string): string {
	return new TextDecoder().decode(
		Uint8Array.from(atob(data), (c) => c.charCodeAt(0)),
	);
}

// abciQuery returns the response Data as text, or null when there is none.
export async function abciQuery(
	path: string,
	data?: string,
	remote: string | null = rpcURL(),
): Promise<string | null> {
	if (!remote) throw new ChainError("no RPC endpoint on this page");
	let url = `${remote}/abci_query?path=${encodeURIComponent(path)}`;
	if (data !== undefined) url += `&data=0x${toHex(data)}`;

	let body: AbciEnvelope;
	try {
		const response = await fetch(url);
		if (!response.ok) throw new ChainError(`${path}: HTTP ${response.status}`);
		body = (await response.json()) as AbciEnvelope;
	} catch (err) {
		if (err instanceof ChainError) throw err;
		throw new ChainError(`${path}: ${String(err)}`);
	}

	const base = body?.result?.response?.ResponseBase;
	if (!base) throw new ChainError(`${path}: malformed response`);
	if (base.Error) {
		const value = base.Error.value ?? "query failed";
		throw new ChainError(`${path}: ${value}`, value);
	}
	return base.Data ? fromBase64(base.Data) : null;
}

// qevalJSON evaluates expr and returns its first result as a string.
export async function qevalJSON(expr: string): Promise<string> {
	const text = await abciQuery("vm/qeval_json", expr);
	try {
		const value = JSON.parse(text ?? "")?.results?.[0]?.V?.value;
		if (typeof value === "string") return value;
	} catch {
		// fall through
	}
	throw new ChainError(`vm/qeval_json: no string result for ${expr}`);
}
