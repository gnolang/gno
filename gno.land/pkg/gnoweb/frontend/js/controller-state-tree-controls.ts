import { BaseController } from "./controller.js";

// StateTreeControlsController is a single-button toggle that flips
// every <details> in the Tree view between expanded and collapsed.
// The button itself owns the state via `aria-pressed`; the icon
// rotates and the label swaps via CSS + a target span.
//
// Why one button instead of two? "Expand all" and "Collapse all"
// are mutually exclusive intents — combining them halves the
// header chrome and removes the moment of "wait, which one am I
// in right now?" by reading the current label.
//
// The actual toggle of `<details open>` cascades a `toggle` event
// per element, which state-tree picks up to persist localStorage.
// Same path as user-driven clicks — one source of persistence
// truth, two surfaces (manual + bulk).
export class StateTreeControlsController extends BaseController {
	protected connect(): void {
		// Nothing to set up — button uses data-action.
	}

	public toggleAll(): void {
		const expand = this.element.getAttribute("aria-pressed") !== "true";
		for (const d of this.allTreeDetails()) {
			if (d.open !== expand) d.open = expand;
		}
		this.element.setAttribute("aria-pressed", String(expand));
		const label = this.getTarget("label");
		if (label) label.textContent = expand ? "Collapse all" : "Expand all";
	}

	private allTreeDetails(): NodeListOf<HTMLDetailsElement> {
		// Scope to the Tree view's <details> only — the page also
		// has unrelated <details> (history disclosure in header,
		// Pretty branch rows). Selecting under .view-tree avoids
		// hijacking those.
		return document.querySelectorAll<HTMLDetailsElement>(
			".b-state-explorer .view-tree details",
		);
	}
}
