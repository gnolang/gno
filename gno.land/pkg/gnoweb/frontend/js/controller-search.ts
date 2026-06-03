import { BaseController, debounce } from "./controller.js";

// Agnostic substring filter primitive. Mount on an `<input>`:
//   data-controller="search"
//   data-search-items-value="<CSS selector>"   (required)
//   data-search-attribute-value="<attr>"       (default: data-name)
//   data-action="input->search#filter"
// Toggles `u-hidden` on non-matching items; no events emitted.
export class SearchController extends BaseController {
	private declare items: HTMLElement[];
	private declare attribute: string;

	protected connect(): void {
		this.items = [];
		this.attribute = "data-name";

		const itemsSelector = this.getValue("items");
		if (!itemsSelector) {
			console.warn("[search] missing required data-search-items-value");
			return;
		}
		const customAttr = this.getValue("attribute");
		if (customAttr) this.attribute = customAttr;

		try {
			this.items = Array.from(
				document.querySelectorAll<HTMLElement>(itemsSelector),
			);
		} catch (err) {
			console.warn(
				`[search] invalid data-search-items-value "${itemsSelector}":`,
				err,
			);
		}
	}

	public filter(event: Event): void {
		const input = event.target;
		if (!(input instanceof HTMLInputElement)) return;
		this.applyFilter(input.value);
	}

	private applyFilter = debounce((value: string): void => {
		const q = value.trim().toLowerCase();
		for (const el of this.items) {
			const v = (el.getAttribute(this.attribute) || "").toLowerCase();
			const match = q === "" || v.includes(q);
			// `u-hidden` over the HTML `hidden` attr: authored `display:*`
			// rules outrank the user-agent `[hidden] { display: none }`.
			el.classList.toggle("u-hidden", !match);
		}
	}, 100);
}
