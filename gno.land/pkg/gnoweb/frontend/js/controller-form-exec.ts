import { BaseController } from "./controller.js";

export class FormExecController extends BaseController {
	protected connect(): void {
		this.initializeDOM({});

		// Find the form element within this controller's scope
		// The form should be either the element itself or a descendant
		const form =
			this.element instanceof HTMLFormElement
				? this.element
				: this.element.querySelector("form");

		if (form) {
			// Listen for submit events
			form.addEventListener("submit", this._handleSubmit.bind(this));
		}
	}

	// Handle form submission
	private _handleSubmit(event: Event): void {
		// Prevent the form from submitting - Extensions should handle the submission
		event.preventDefault();
		event.stopPropagation();

		const actionFunction = this.getTarget("command");
		if (actionFunction) {
			actionFunction.classList.remove("u-hidden");
		}
	}
}
