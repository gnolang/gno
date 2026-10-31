// Caps the text walk on huge realm pages.
export const MAX_TEXT_NODES = 5000;

// linkIsSelf reports whether href points at the connected identity.
export function linkIsSelf(
	href: string,
	base: string,
	address: string,
	username?: string,
): boolean {
	let url: URL;
	let path: string;
	try {
		url = new URL(href, base);
		path = decodeURIComponent(url.pathname);
	} catch {
		return false;
	}
	if (username && url.origin === new URL(base).origin) {
		if (path === `/u/${username}`) return true;
	}
	// The path names who it points at; a query string is page state.
	return path.includes(address);
}
