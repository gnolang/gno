import { BaseController } from "./controller.js";

export class FormExecController extends BaseController {
	protected connect(): void {
		this.initializeDOM({});

		const form =
			this.element instanceof HTMLFormElement
				? this.element
				: this.element.querySelector("form");

		if (form) {
			form.addEventListener("submit", this._handleSubmit.bind(this));
		}
	}

	// Handle form submission
	private _handleSubmit(event: Event): void {
		// Prevent the form from submitting - Extensions should handle the submission.
		// stopPropagation here is load-bearing for analytics.ts: the submit_action
		// listener must run on the capture phase to fire before this point.
		event.preventDefault();
		event.stopPropagation();

		const actionFunction = this.getTarget("command");
		if (actionFunction) {
			actionFunction.classList.remove("u-hidden");
		}
	}
}
