import { BaseController, debounce } from "./controller.js";

// StateSearchController filters top-level Pretty cards in place by
// name (case-insensitive substring). Tree view doesn't expose a
// search input — the filter would only reach top-level rows and
// would mislead the user about its scope.
//
// No URL state, no fancy routing — purely a "find as you type"
// experience. Reload clears the filter, which is the expected UX
// for an ephemeral search.
export class StateSearchController extends BaseController {
	protected connect(): void {
		// Nothing to set up — the input listens via data-action.
	}

	// filter is the input-event handler stamped on the search input.
	// Defined as a regular method (not a class field) so it lives on
	// the prototype and is reachable from BaseController's setupActions
	// — class fields initialise *after* super() returns, which is
	// after init()/setupActions() has already run.
	public filter(event: Event): void {
		const input = event.target as HTMLInputElement;
		this.applyFilter(input.value);
	}

	// Debounced to avoid thrashing the DOM on every keystroke. O(N)
	// over Pretty cards is cheap, but no need to run it 50×/sec.
	private applyFilter = debounce((value: string): void => {
		const q = value.trim().toLowerCase();
		const items = document.querySelectorAll<HTMLElement>(
			".b-state-explorer [data-name]",
		);
		for (const el of items) {
			const name = (el.getAttribute("data-name") || "").toLowerCase();
			const match = q === "" || name.includes(q);
			el.hidden = !match;
		}
	}, 100);
}
