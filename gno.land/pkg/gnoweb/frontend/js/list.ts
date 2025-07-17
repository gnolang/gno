import { debounce } from "./utils";

type ListItem = {
	element: HTMLElement;
	title: string;
	type: string;
};

class List {
	private DOM: {
		el: Element | null;
		range: HTMLElement | null;
		searchBar: HTMLInputElement | null;
		orderOption: HTMLElement | null;
		packagesCount: HTMLElement | null;
		realmsCount: HTMLElement | null;
		pureCount: HTMLElement | null;
	};

	private static SELECTORS = {
		range: ".js-list-range",
		searchBar: ".js-list-searchbar",
		orderOption: ".js-list-order-filter",
		itemTitle: ".js-list-range-title",
		packagesCount: ".js-list-packages-count",
		realmsCount: ".js-list-realms-count",
		pureCount: ".js-list-pure-count",
	};

	private static STORAGE_KEY = "gno_display_mode";
	private static LOADING_CLASS = "is-loading";

	private items: Array<ListItem> = [];

	// Cache
	private currentFilter: string = "";
	private sortedItems: Array<ListItem> = [];

	constructor(el: Element | null) {
		if (!el) {
			console.warn("No element provided");
			return;
		}

		this.DOM = {
			el,
			range: el.querySelector<HTMLElement>(List.SELECTORS.range),
			searchBar: el.querySelector<HTMLInputElement>(List.SELECTORS.searchBar),
			orderOption: el.querySelector<HTMLElement>(List.SELECTORS.orderOption),
			packagesCount: el.querySelector<HTMLElement>(
				List.SELECTORS.packagesCount,
			),
			realmsCount: el.querySelector<HTMLElement>(List.SELECTORS.realmsCount),
			pureCount: el.querySelector<HTMLElement>(List.SELECTORS.pureCount),
		};

		if (!this.DOM.range) {
			console.warn("Range element not found");
			return;
		}

		// Store initial items with their titles and types
		this.items = Array.from(this.DOM.range.children).map((element) => {
			const titleElement = (element as HTMLElement).querySelector(
				List.SELECTORS.itemTitle,
			);
			const type = (element as HTMLElement).dataset.type || "";

			return {
				element: element as HTMLElement,
				title: (titleElement?.textContent || "").toLowerCase(),
				type,
			};
		});

		this.sortedItems = [...this.items];
		this.restoreDisplayMode();
		this.bindEvents();

		// Remove loading state after a small delay to ensure styles are applied
		requestAnimationFrame(() => {
			el.classList.remove(List.LOADING_CLASS);
		});
	}

	private bindEvents(): void {
		const { searchBar, orderOption } = this.DOM;

		// Handle order change
		orderOption?.addEventListener("change", (e) => {
			const target = e.target as HTMLInputElement;

			// event delegation
			if (target.matches('input[name="order-mode"]')) {
				this.sortItems();
				this.updateDOM();
			}
		});

		// Handle display mode change
		this.DOM.el?.addEventListener("change", (e) => {
			const target = e.target as HTMLInputElement;
			if (target.matches('input[name="display-mode"]')) {
				localStorage.setItem(List.STORAGE_KEY, target.value);
			}
		});

		searchBar?.addEventListener("input", (e) => {
			const target = e.target as HTMLInputElement;
			this.debouncedSearch(target.value);
		});
	}

	private restoreDisplayMode(): void {
		const savedMode = localStorage.getItem(List.STORAGE_KEY);
		if (!savedMode) return;

		const input = this.DOM.el?.querySelector<HTMLInputElement>(
			`input[name="display-mode"][value="${savedMode}"]`,
		);
		if (input && !input.checked) {
			input.checked = true;
		}
	}

	// Debounce search to avoid too many calls
	private debouncedSearch = debounce((value: string) => {
		this.currentFilter = value.toLowerCase();
		this.updateDOM();
	}, 150);

	// Sort items
	private sortItems(): void {
		this.sortedItems.reverse();
	}

	// Filter items
	private filterItems(): Array<ListItem> {
		if (!this.currentFilter) return this.sortedItems;

		return this.sortedItems.filter((item) =>
			item.title.includes(this.currentFilter),
		);
	}

	// Update counts based on filtered items
	private updateCounts(items: Array<ListItem>): void {
		const { packagesCount, realmsCount, pureCount } = this.DOM;

		const counts = items.reduce(
			(acc, item) => {
				acc[item.type] = (acc[item.type] || 0) + 1;
				return acc;
			},
			{} as Record<string, number>,
		);

		const realmCountValue = counts.realm || 0;
		const pureCountValue = counts.pure || 0;
		const totalPackages = realmCountValue + pureCountValue;

		if (packagesCount) {
			packagesCount.textContent = totalPackages.toString();
		}
		if (realmsCount) {
			realmsCount.textContent = realmCountValue.toString();
		}
		if (pureCount) {
			pureCount.textContent = pureCountValue.toString();
		}
	}

	// Update DOM with items
	private updateDOM(): void {
		const { range } = this.DOM;
		if (!range) return;

		const filteredItems = this.filterItems();

		// Create a document fragment for batch DOM updates
		const fragment = document.createDocumentFragment();
		filteredItems.forEach((item) => fragment.appendChild(item.element));

		// Clear and update range with a single reflow
		range.textContent = "";
		range.appendChild(fragment);

		// Update counts based on filtered items
		this.updateCounts(filteredItems);
	}
}

export default (el: HTMLElement | null) => new List(el);
