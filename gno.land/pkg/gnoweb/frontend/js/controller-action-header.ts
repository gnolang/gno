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
			this._restoreAddress();
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

	// restore the address from localStorage
	private _restoreAddress(): void {
		const addressInput = this.getTarget("address") as HTMLInputElement;
		this.restoreValue("actionAddressInput", addressInput, (value) => {
			// Dispatch event for other controllers to listen
			this.dispatch("address:changed", { address: value });
		});
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
