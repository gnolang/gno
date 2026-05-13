import { BaseController } from "./controller.js";

// Matches Amino object IDs like "ff61a23bc5:12" or ":1" (cycle-break refs).
// Format: `<hex>:<uint>` per ObjectID.MarshalAmino in gnovm/pkg/gnolang/ownership.go.
const OID_PATTERN = /^[a-f0-9]*:\d+$/i;

export class SearchbarController extends BaseController {
	protected connect(): void {
		this.initializeDOM({
			input: this.getTarget("input"),
			breadcrumb: this.getTarget("breadcrumb"),
		});
	}

	public searchUrl(e: Event): void {
		e.preventDefault();

		const inputElement = this.getDOMElement("input");
		if (!(inputElement instanceof HTMLInputElement)) {
			this.warn("input target missing or wrong type");
			return;
		}
		let url = inputElement.value.trim();
		if (!url) {
			this.warn("empty URL");
			return;
		}

		// OID-shaped input redirects to the state view for that object,
		// preserving ?height=N so time-travel survives the jump.
		if (OID_PATTERN.test(url) && !url.startsWith("/")) {
			const realmPath = this.currentRealmPath();
			if (realmPath) {
				const h = new URLSearchParams(location.search).get("height");
				const pin = h && /^\d+$/.test(h) ? `&height=${h}` : "";
				location.href = `${realmPath}$state&oid=${encodeURIComponent(url)}${pin}`;
				return;
			}
		}

		if (!/^https?:\/\//i.test(url)) {
			const baseUrl = window.location.origin;
			url = `${baseUrl}${url.startsWith("/") ? "" : "/"}${url}`;
		}

		try {
			window.location.href = new URL(url).href;
		} catch (_error) {
			this.warn(`invalid URL: ${url}`);
		}
	}

	private currentRealmPath(): string | null {
		const match = window.location.pathname.match(/^(\/r\/[^$]+)/);
		return match ? match[1] : null;
	}
}
