// Browser globals for unit tests. Importing this module installs a default
// DOM, because some modules (wallet-discovery.ts) touch window on load.

export interface FakeDOM {
	storage: Map<string, string>;
	metas: Record<string, string>;
}

export function installDOM(metas: Record<string, string> = {}): FakeDOM {
	const storage = new Map<string, string>();
	const localStorage = {
		getItem: (key: string) => storage.get(key) ?? null,
		setItem: (key: string, value: string) => {
			storage.set(key, value);
		},
		removeItem: (key: string) => {
			storage.delete(key);
		},
	};
	const win = Object.assign(new EventTarget(), { localStorage });
	const doc = Object.assign(new EventTarget(), {
		querySelector: (selector: string) => {
			const match = /^meta\[name="([^"]+)"\]$/.exec(selector);
			const content = match ? metas[match[1]] : undefined;
			if (content === undefined) return null;
			return {
				getAttribute: (name: string) => (name === "content" ? content : null),
			};
		},
	});
	Object.assign(globalThis, { window: win, document: doc });
	return { storage, metas };
}

export interface FakeResponse {
	status?: number;
	body: unknown;
}

// stubFetch replaces fetch with a handler and returns the requested URLs.
export function stubFetch(
	handler: (url: string) => FakeResponse | Error,
): string[] {
	const calls: string[] = [];
	globalThis.fetch = (async (input: string | URL | Request) => {
		const url = input instanceof Request ? input.url : String(input);
		calls.push(url);
		const res = handler(url);
		if (res instanceof Error) throw res;
		return new Response(JSON.stringify(res.body), {
			status: res.status ?? 200,
		});
	}) as typeof fetch;
	return calls;
}

// abciBody is an abci_query JSON-RPC answer: base64 Data, or a node Error.
export function abciBody(data: string | null, error?: string): unknown {
	return {
		jsonrpc: "2.0",
		id: "",
		result: {
			response: {
				ResponseBase: {
					Error: error ? { "@type": "/abci.StringError", value: error } : null,
					Data: data === null ? null : btoa(data),
					Events: null,
					Log: "",
					Info: "",
				},
			},
		},
	};
}

installDOM();
