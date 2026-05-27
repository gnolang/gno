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
		const raw = inputElement.value.trim();
		if (!raw) {
			this.warn("empty URL");
			return;
		}

		// OID-shaped input redirects to the state view for that object.
		if (OID_PATTERN.test(raw) && !raw.startsWith("/")) {
			const realmPath = this.currentRealmPath();
			if (realmPath) {
				location.href = `${realmPath}$state&oid=${encodeURIComponent(raw)}`;
				return;
			}
		}

		// Otherwise: strip a leading gno.land host + resolve against the
		// current origin so realm paths copied from anywhere land locally.
		const target = SearchbarController.resolveTarget(raw);
		if (target === null) {
			this.warn(`invalid URL: ${raw}`);
			return;
		}
		window.location.href = target;
	}

	// resolveTarget strips a leading `gno.land` host (with or without scheme)
	// so realm paths copied from anywhere resolve locally; non-`gno.land`
	// absolute URLs pass through, and relatives resolve against the origin.
	static resolveTarget(input: string): string | null {
		const stripped = input.replace(
			/^(?:https?:\/\/)?gno\.land(?=\/|$|\?|#)/i,
			"",
		);
		const url = URL.parse(stripped, window.location.origin);
		if (!url) {
			console.error("SearchBarController: Invalid URL.");
			return null;
		}
		return url.href;
	}

	private currentRealmPath(): string | null {
		const match = window.location.pathname.match(/^(\/r\/[^$]+)/);
		return match ? match[1] : null;
	}
}
