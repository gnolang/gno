import { BaseController, debounce } from "./controller.js";

// Agnostic substring filter primitive. Mount on an `<input>`:
//   data-controller="search"
//   data-search-items-value="<CSS selector>"   (required)
//   data-search-attribute-value="<attr>"       (default: data-name)
//   data-action="input->search#filter"
// After each pass, dispatches a bubbling `search:filter` CustomEvent
// with `{ query, matchCount, totalCount }` for downstream integrations
// (e.g. a counter badge that reflects the filtered total).
export class SearchController extends BaseController {
	// `declare` (no init): field exists at runtime but no initializer
	// emitted. Critical — BaseController's constructor calls init()
	// (→ connect()) BEFORE derived class-field initialisers run, so
	// any `= default` here would clobber connect-time assignments.
	private declare items: HTMLElement[];
	private declare attribute: string;

	protected connect(): void {
		this.items = [];
		this.attribute = "data-name";

		const itemsSelector = this.getValue("items");
		if (!itemsSelector) {
			console.warn(
				"SearchController: missing required data-search-items-value",
			);
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
				`SearchController: invalid data-search-items-value "${itemsSelector}":`,
				err,
			);
		}
	}

	// Method (not class field) so it lives on the prototype and
	// BaseController's setupActions can bind it.
	public filter(event: Event): void {
		const input = event.target as HTMLInputElement;
		this.applyFilter(input.value);
	}

	private applyFilter = debounce((value: string): void => {
		const q = value.trim().toLowerCase();
		let matchCount = 0;
		for (const el of this.items) {
			const v = (el.getAttribute(this.attribute) || "").toLowerCase();
			const match = q === "" || v.includes(q);
			// Use the `u-hidden` utility (`display: none`) rather than
			// the HTML `hidden` attribute — `[hidden]` defaults to
			// `display: none` from the user-agent stylesheet, but any
			// authored `display: <whatever>` rule on the same element
			// (e.g. `.b-state-decl { display: flex }`) overrides it.
			// Class-based hide is more robust against cascade collisions.
			el.classList.toggle("u-hidden", !match);
			if (match) matchCount++;
		}
		this.element.dispatchEvent(
			new CustomEvent("search:filter", {
				bubbles: true,
				detail: { query: q, matchCount, totalCount: this.items.length },
			}),
		);
	}, 100);
}
