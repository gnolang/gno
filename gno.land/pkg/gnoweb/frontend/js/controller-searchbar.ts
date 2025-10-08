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
		let url = inputElement?.value.trim();

		if (url) {
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
		} else {
			console.error("SearchBarController: Please enter a URL to search.");
		}
	}
}
