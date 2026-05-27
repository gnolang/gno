import { BaseController } from "./controller.js";

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
		const raw = inputElement?.value.trim();

		if (!raw) {
			console.error("SearchBarController: Please enter a URL to search.");
			return;
		}

		const target = SearchbarController.resolveTarget(raw);
		if (target === null) return;
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
}
