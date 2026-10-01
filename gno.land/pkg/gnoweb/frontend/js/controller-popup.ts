import { BaseController } from "./controller.js";

// PopupController lets the keyboard open and close a CSS-only popup. Its
// toggle is a <label role="button"> so the popup still works without JS, but
// a label only reacts to clicks: Enter and Space now click it too.
export class PopupController extends BaseController {
	protected connect(): void {}

	public key(event: Event): void {
		const e = event as KeyboardEvent;
		if (e.repeat || e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;
		if (e.key !== "Enter" && e.key !== " ") return;
		event.preventDefault();
		this.element.click();
	}
}
