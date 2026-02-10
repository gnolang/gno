import { BaseController, debounce } from "./controller.js";

type ListItem = {
	element: HTMLElement;
	title: string;
	type: string;
};

export class ListController extends BaseController {
	private static STORAGE_KEY = "gno_display_mode";
	private static LOADING_CLASS = "u-is-loading";

	declare currentFilter: string;
	declare items: Array<ListItem>;
	declare sortedItems: Array<ListItem>;

	protected connect(): void {
		this.initializeDOM({
			range: this.getTarget("range"),
			packagesCount: this.getTarget("packages-count"),
			realmsCount: this.getTarget("realms-count"),
			pureCount: this.getTarget("pure-count"),
		});

		this.currentFilter = "";

		if (!this.getDOMElement("range")) {
			console.warn("ListController: Range element not found");
			return;
		}

		// Store initial items with their titles and types
		this._initializeItems();
		this._restoreDisplayMode();

		// Remove loading state after a small delay to ensure styles are applied
		requestAnimationFrame(() => {
			this.element?.classList.remove(ListController.LOADING_CLASS);
		});
	}

	// initialize the items
	private _initializeItems(): void {
		const rangeElement = this.getDOMElement("range");
		if (!rangeElement) return;

		// Get all range-title elements
		const titleElements = this.getTargets("range-title");

		// map the items to capture the title and type (cached)
		this.items = Array.from(rangeElement.children).map((element, index) => {
			const titleElement = titleElements[index];
			const type = this.getValue("type", element as HTMLElement);

			return {
				element: element as HTMLElement,
				title: (titleElement?.textContent || "").toLowerCase(),
				type,
			};
		});

		this.sortedItems = [...this.items];
	}

	// Restore display mode (grid or list)
	private _restoreDisplayMode(): void {
		const savedMode = localStorage.getItem(ListController.STORAGE_KEY);
		if (!savedMode) return;

		// complex selector to find the input element (not using getTarget)
		const input = this.element?.querySelector<HTMLInputElement>(
			`input[name="display-mode"][value="${savedMode}"]`,
		);

		if (input && !input.checked) {
			input.checked = true;
		}
	}

	// Update counts based on filtered items
	private _updateCounts(items: Array<ListItem>): void {
		const packagesCount = this.getDOMElement("packagesCount");
		const realmsCount = this.getDOMElement("realmsCount");
		const pureCount = this.getDOMElement("pureCount");

		// count the items by type
		const counts = items.reduce(
			(acc, item) => {
				acc[item.type] = (acc[item.type] || 0) + 1;
				return acc;
			},
			{} as Record<string, number>,
		);

		// get the counts for each type
		const realmCountValue = counts.realm || 0;
		const pureCountValue = counts.pure || 0;
		const totalPackages = realmCountValue + pureCountValue;

		// set the text content of the counts
		if (packagesCount) packagesCount.textContent = totalPackages.toString();
		if (realmsCount) realmsCount.textContent = realmCountValue.toString();
		if (pureCount) pureCount.textContent = pureCountValue.toString();
	}

	// Update DOM with items
	private _updateDOM(): void {
		const range = this.getDOMElement("range");
		if (!range) return;

		// filter the items based on the current filter
		const filteredItems = this._filterItems();

		// Create a document fragment for batch DOM updates
		const fragment = document.createDocumentFragment();
		filteredItems.forEach((item) => fragment.appendChild(item.element));

		// Clear and update range with a single reflow
		range.textContent = "";
		range.appendChild(fragment);

		// Update counts based on filtered items
		this._updateCounts(filteredItems);
	}

	// Debounce search to avoid too many calls
	private _debouncedSearch = debounce((value: string) => {
		this.currentFilter = value.toLowerCase();
		this._updateDOM();
	}, 150);

	// Sort items
	private _sortItems(): void {
		this.sortedItems.reverse();
	}

	// Filter items
	private _filterItems(): Array<ListItem> {
		if (!this.currentFilter) return this.sortedItems;

		// filter the items based on the current filter
		return this.sortedItems.filter((item) =>
			item.title.includes(this.currentFilter),
		);
	}

	// ACTIONS
	// handle search from input in list
	public search(e: Event): void {
		const target = e.target as HTMLInputElement;
		this._debouncedSearch(target.value);
	}

	// handle order change to switch between asc and desc
	public orderChange(_e: Event): void {
		this._sortItems();
		this._updateDOM();
	}

	// handle display mode change to switch between grid and list
	public displayModeChange(e: Event): void {
		const target = e.target as HTMLInputElement;
		localStorage.setItem(ListController.STORAGE_KEY, target.value);
	}
}
