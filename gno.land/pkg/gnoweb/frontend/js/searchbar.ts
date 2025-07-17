class SearchBar {
	private DOM: {
		el: HTMLElement | null;
		inputSearch: HTMLInputElement | null;
		breadcrumb: HTMLElement | null;
	};

	private baseUrl: string;

	private static SELECTORS = {
		container: ".js-header-searchbar",
		inputSearch: "[data-role='header-input-search']",
		breadcrumb: "[data-role='header-breadcrumb-search']",
	};

	constructor() {
		this.DOM = {
			el: document.querySelector<HTMLElement>(SearchBar.SELECTORS.container),
			inputSearch: null,
			breadcrumb: null,
		};

		this.baseUrl = window.location.origin;

		if (this.DOM.el) {
			this.init();
		} else {
			console.warn("SearchBar: Main container not found.");
		}
	}

	private init(): void {
		const { el } = this.DOM;

		this.DOM.inputSearch =
			el?.querySelector<HTMLInputElement>(SearchBar.SELECTORS.inputSearch) ??
			null;
		this.DOM.breadcrumb =
			el?.querySelector<HTMLElement>(SearchBar.SELECTORS.breadcrumb) ?? null;

		if (!this.DOM.inputSearch) {
			console.warn("SearchBar: Input element for search not found.");
		}

		this.bindEvents();
	}

	private bindEvents(): void {
		this.DOM.el?.addEventListener("submit", (e) => {
			e.preventDefault();
			this.searchUrl();
		});
	}

	public searchUrl(): void {
		const input = this.DOM.inputSearch?.value.trim();

		if (input) {
			let url = input;

			// Check if the URL has a proper scheme
			if (!/^https?:\/\//i.test(url)) {
				url = `${this.baseUrl}${url.startsWith("/") ? "" : "/"}${url}`;
			}

			try {
				window.location.href = new URL(url).href;
			} catch (_error) {
				console.error(
					"SearchBar: Invalid URL. Please enter a valid URL starting with http:// or https://.",
				);
			}
		} else {
			console.error("SearchBar: Please enter a URL to search.");
		}
	}
}

export default () => new SearchBar();
