import { truncate } from "../../feature/connect/frontend/session.js";

export const SIGNER_UNAVAILABLE = "signer_unavailable";

// helpFunc is the function a help URL targets: `$help&func=X` in the path,
// or `?func=X` on the static fixture.
export function helpFunc(url: URL): string {
	const at = url.pathname.indexOf("$");
	if (at >= 0) {
		const func = new URLSearchParams(url.pathname.slice(at + 1)).get("func");
		if (func) return func;
	}
	return url.searchParams.get("func") ?? "";
}

// outcomeMessage is the notice for funcName, or null when the URL has none.
export function outcomeMessage(
	href: string,
	funcName: string,
	address: string | null,
): string | null {
	const url = new URL(href);
	if (url.searchParams.get("status") !== "error") return null;
	if (url.searchParams.get("code") !== SIGNER_UNAVAILABLE) return null;
	if (!funcName || helpFunc(url) !== funcName) return null;
	return address
		? `Your wallet isn't on ${truncate(address)}, the account you're connected as. Switch to it in the wallet, or reconnect.`
		: "Your wallet isn't on the account you connected with. Switch to it in the wallet, or reconnect.";
}

export function withoutOutcome(href: string): string {
	const url = new URL(href);
	url.searchParams.delete("status");
	url.searchParams.delete("code");
	return url.toString();
}
