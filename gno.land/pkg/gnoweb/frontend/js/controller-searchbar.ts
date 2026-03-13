import { BaseController } from "./controller.js";

// Matches Amino object IDs like "a]0000000001" or "ff61a23bc5:12"
const OID_PATTERN = /^[a-f0-9\]:.]+$/i;

export class SearchbarController extends BaseController {
	protected connect(): void {
		this.initializeDOM({
			input: this.getTarget("input"),
			breadcrumb: this.getTarget("breadcrumb"),
		});
	}

	public searchUrl(e: Event): void {
		e.preventDefault();

		const inputElement = this.getDOMElement("input") as HTMLInputElement;
		let url = inputElement?.value.trim();

		if (!url) {
			console.error("SearchBarController: Please enter a URL to search.");
			return;
		}

		// Detect object IDs and redirect to state view
		if (OID_PATTERN.test(url) && !url.startsWith("/")) {
			const realmPath = this._currentRealmPath();
			if (realmPath) {
				window.location.href = `${realmPath}$state&oid=${encodeURIComponent(url)}`;
				return;
			}
		}

		// Check if the URL has a proper scheme
		if (!/^https?:\/\//i.test(url)) {
			const baseUrl = window.location.origin;
			url = `${baseUrl}${url.startsWith("/") ? "" : "/"}${url}`;
		}

		try {
			window.location.href = new URL(url).href;
		} catch (_error) {
			console.error(
				"SearchBarController: Invalid URL. Please enter a valid URL starting with http:// or https://.",
			);
		}
	}

	private _currentRealmPath(): string | null {
		const path = window.location.pathname;
		// Match realm paths like /r/demo/tamagotchi
		const match = path.match(/^(\/r\/[^$]+)/);
		return match ? match[1] : null;
	}
}
