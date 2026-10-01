import { BaseController } from "./controller.js";

// PopupController lets the keyboard open and close a CSS-only popup. Its
// toggle is a <label role="button"> so the popup still works without JS, but
// a label only reacts to clicks, and a click moves focus to its hidden
// checkbox: Enter and Space toggle the checkbox instead, leaving focus on the
// label. On the opener (the label with aria-controls), aria-expanded follows
// the checkbox, and focus left on the checkbox or in the closed popup
// returns to the opener.
export class PopupController extends BaseController {
	protected connect(): void {
		this.input()?.addEventListener("change", () => this.sync());
		this.sync();
	}

	public key(event: Event): void {
		const e = event as KeyboardEvent;
		if (e.repeat || e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;
		if (e.key !== "Enter" && e.key !== " ") return;
		event.preventDefault();
		const input = this.input();
		if (!input) return;
		input.checked = !input.checked;
		// Bubbles so the analytics listener on the checkbox still fires.
		input.dispatchEvent(new Event("change", { bubbles: true }));
	}

	private input(): HTMLInputElement | null {
		const id = (this.element as HTMLLabelElement).htmlFor;
		return id ? (document.getElementById(id) as HTMLInputElement | null) : null;
	}

	private sync(): void {
		const popupId = this.element.getAttribute("aria-controls");
		const input = this.input();
		if (!popupId || !input) return;
		this.element.setAttribute("aria-expanded", String(input.checked));
		const active = document.activeElement;
		const popup = document.getElementById(popupId);
		if (active === input || (!input.checked && popup?.contains(active))) {
			this.element.focus();
		}
	}
}
