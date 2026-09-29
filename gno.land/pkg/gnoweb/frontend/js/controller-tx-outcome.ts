import { readSession } from "../../feature/connect/frontend/session.js";
import { BaseController } from "./controller.js";
import { outcomeMessage, withoutOutcome } from "./tx-outcome.js";

// TxOutcomeController shows why a transaction came back, for the function it
// sits in. Only signer_unavailable is explained today.
export class TxOutcomeController extends BaseController {
	protected connect(): void {
		const message = outcomeMessage(
			window.location.href,
			this.getValue("func"),
			readSession()?.address ?? null,
		);
		if (!message) return;

		const text = this.getTarget("message");
		if (text) text.textContent = message;
		this.element.hidden = false;

		this.getTarget("reconnect")?.addEventListener("click", () =>
			this.dispatch("connect:pick"),
		);
		this.getTarget("dismiss")?.addEventListener("click", () => {
			window.history.replaceState(
				null,
				"",
				withoutOutcome(window.location.href),
			);
			this.element.hidden = true;
		});
	}
}
