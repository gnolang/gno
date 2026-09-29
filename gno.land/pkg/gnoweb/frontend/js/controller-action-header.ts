import {
	type GnoSession,
	onSessionChange,
	readSession,
} from "../../feature/connect/frontend/session.js";
import { refresh } from "./account.js";
import { BaseController, debounce } from "./controller.js";

// TYPE DEFINITIONS
export const ActionModeValues = {
	Fast: "fast",
	Secure: "secure",
} as const;

export type ActionMode =
	(typeof ActionModeValues)[keyof typeof ActionModeValues];

// CONTROLLER
export class ActionHeaderController extends BaseController {
	protected connect(): void {
		this.on("controllers:ready", () => {
			this._restoreMode();
			this._applySession(readSession());
		});
		onSessionChange((session) => this._applySession(session));
		// Back from the terminal: the sequence has likely moved. One owner
		// refreshes, so N function blocks share one request.
		document.addEventListener("visibilitychange", () => {
			const input = this.getTarget("address") as HTMLInputElement | null;
			if (document.visibilityState !== "visible" || !input?.value) return;
			refresh(input.value.trim());
			this.dispatch("address:changed", { address: input.value });
		});
	}

	// restore a value from localStorage
	private restoreValue(
		storageKey: string,
		inputElement: HTMLInputElement | HTMLSelectElement | null,
		updateCallback: (value: string) => void,
	): void {
		if (inputElement) {
			const storedValue = localStorage.getItem(storageKey);
			if (storedValue) {
				inputElement.value = storedValue;
				updateCallback(storedValue);
			}
		}
	}

	// restore the mode from localStorage
	private _restoreMode(): void {
		const cmdModeSelect = this.getTarget("mode") as HTMLSelectElement;
		this.restoreValue("actionCmdMode", cmdModeSelect, (value) => {
			// Dispatch event for other controllers to listen
			if (
				value === ActionModeValues.Fast ||
				value === ActionModeValues.Secure
			) {
				this.dispatch("mode:changed", { mode: value as ActionMode });
			}
		});
	}

	// Connected: the session address, still editable. Disconnected: the
	// manually typed address, as before.
	private _applySession(session: GnoSession | null): void {
		const input = this.getTarget("address") as HTMLInputElement | null;
		if (!input) return;
		const address =
			session?.address ?? localStorage.getItem("actionAddressInput") ?? "";
		input.value = address;
		this.dispatch("address:changed", { address });
	}

	// debounced address update
	private _debouncedAddressUpdate = debounce(
		(addressInput: HTMLInputElement) => {
			const address = addressInput.value;
			localStorage.setItem("actionAddressInput", address);
			this.dispatch("address:changed", { address });
		},
		50,
	);

	// DOM ACTIONS
	// update the mode (DOM action)
	public updateMode(event: Event): void {
		const target = event.target as HTMLSelectElement;
		const mode = target.value as ActionMode;
		localStorage.setItem("actionCmdMode", mode);

		// Dispatch event for other controllers to listen
		this.dispatch("mode:changed", { mode });
	}

	// update the address (debounced - DOM action)
	public updateAddress(event: Event): void {
		const target = event.target as HTMLInputElement;
		this._debouncedAddressUpdate(target);
	}
}
