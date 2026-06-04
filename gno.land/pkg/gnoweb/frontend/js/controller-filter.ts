import { BaseController, debounce } from "./controller.js";

// Agnostic substring filter primitive. Mount on an `<input>`:
//   data-controller="filter"
//   data-filter-items-value="<CSS selector>"   (required)
//   data-filter-attribute-value="<attr>"       (default: data-name)
//   data-action="input->filter#filter"
// Toggles `u-hidden` on non-matching items; no events emitted.
export class FilterController extends BaseController {
	private declare items: HTMLElement[];
	private declare attribute: string;

	protected connect(): void {
		this.items = [];
		this.attribute = "data-name";

		const itemsSelector = this.getValue("items");
		if (!itemsSelector) {
			console.warn("[filter] missing required data-filter-items-value");
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
				`[filter] invalid data-filter-items-value "${itemsSelector}":`,
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
