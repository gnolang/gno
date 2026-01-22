import {
	BaseController,
	debounce,
	escapeShellSpecialChars,
} from "./controller.js";
import type { ActionMode } from "./controller-action-header.js";

export class ActionFunctionController extends BaseController {
	protected sendValue: string | null = null;
	declare _funcName: string | null;
	declare _pkgPath: string | null;
	declare _paramInputsCache: HTMLInputElement[];

	// Cached params inputs
	private get _paramInputs(): HTMLInputElement[] {
		if (!this._paramInputsCache) {
			this._paramInputsCache = this.getTargets(
				"param-input",
			) as HTMLInputElement[];
		}
		return this._paramInputsCache;
	}

	protected connect(): void {
		this.initializeDOM({
			"send-code": this.getTargets("send-code"),
		});

		this._funcName = this.getValue("name") || null;
		this._pkgPath = this.getValue("pkgpath") || null;

		this._initializeArgs();
		this._listenForEvents();

		// Some functions may have no params, or all params already have values
		this._updateQEvalResult();

		// Dispatch initial params state for wallet integration
		this._dispatchParamsChanged();
	}

	// listen for events from action-header controller
	private _listenForEvents(): void {
		// Listen for mode changes from action-header controller
		this.on("mode:changed", (event: Event) => {
			const customEvent = event as CustomEvent;
			const mode: ActionMode = customEvent.detail.mode;
			this._updateAllFunctionsMode(mode);
		});

		// Listen for address changes from action-header controller
		this.on("address:changed", (event: Event) => {
			const customEvent = event as CustomEvent;
			const address = customEvent.detail.address;
			this._updateAllFunctionsAddress(address);
		});
	}

	// update all functions mode
	private _updateAllFunctionsMode(mode: ActionMode): void {
		// Update mode-specific elements within each function scope
		const modeElements = this.getTargets("mode");

		modeElements.forEach((modeElement) => {
			const isVisible = this.getValue("mode", modeElement) === mode;
			modeElement.classList.toggle("u-inline", isVisible);
			modeElement.classList.toggle("u-hidden", !isVisible);
			modeElement.dataset.copyTarget =
				isVisible && this._funcName ? `action-function-${this._funcName}` : "";
		});
	}

	// update all functions address
	private _updateAllFunctionsAddress(address: string): void {
		// Update address elements
		const addressElements = this.getTargets("address");
		addressElements.forEach((addressElement) => {
			addressElement.textContent = address.trim() || "ADDRESS";
		});
	}

	// sanitize the args input
	private _sanitizeArgsInput(input: HTMLInputElement): {
		paramName: string;
		paramValue: string;
	} {
		const paramName = this.getValue("param", input) || "";
		const paramValue = input.value.trim();

		if (!paramName) {
			console.warn("sanitizeArgsInput: param is missing in arg input dataset.");
		}

		return { paramName, paramValue };
	}

	// get current value for a param name (handles checkbox multiple values)
	private _getParamCurrentValue(paramName: string): string {
		// radio or checkbox multiple values
		const inputs = this._paramInputs.filter(
			(inp) => this.getValue("param", inp) === paramName,
		);

		if (!inputs.length) return "";

		const firstInput = inputs[0];

		// Checkbox: join all checked values
		if (firstInput.type === "checkbox") {
			return inputs
				.filter((inp) => inp.checked || inp.defaultChecked)
				.map((inp) => inp.value.trim())
				.join(",");
		}

		// Radio: find checked one
		if (firstInput.type === "radio") {
			const checked = inputs.find((inp) => inp.checked || inp.defaultChecked);
			return checked?.value.trim() || "";
		}

		// Other: return value
		return firstInput.value.trim();
	}

	// initialize the args
	private _initializeArgs(): void {
		// multiple values (radio or checkbox) to be initialized only once
		const processed = new Set<string>();

		// initialize the args
		this._paramInputs.forEach((input) => {
			const paramName = this.getValue("param", input) || "";

			if (!paramName || processed.has(paramName)) return;

			const paramValue = this._getParamCurrentValue(paramName);
			if (paramValue) this._updateArgInDOM(paramName, paramValue);

			processed.add(paramName);
		});
	}

	// debounced update all args and update the qeval result
	private _debouncedUpdateAllArgs = debounce(
		(paramName: string, paramValue: string) => {
			if (paramName) {
				this._updateArgInDOM(paramName, paramValue);
				this._updateQEvalResult();
				this._dispatchParamsChanged();
			}
		},
		50,
	);

	// Dispatch params:changed event uppon parameter updates.
	private _dispatchParamsChanged(): void {
		if (!this._funcName || !this._pkgPath) return;

		const params: Record<string, string> = {};
		const processed = new Set<string>();

		this._paramInputs.forEach((input) => {
			const paramName = this.getValue("param", input) || "";
			if (!paramName || processed.has(paramName)) return;

			params[paramName] = this._getParamCurrentValue(paramName);
			processed.add(paramName);
		});

		this.dispatch("params:changed", {
			pkgPath: this._pkgPath,
			funcName: this._funcName,
			params: params,
			send: this.sendValue || undefined,
		});
	}

	// push args in DOM (in func code)
	private _updateArgInDOM(paramName: string, paramValue: string): void {
		const escapedValue = escapeShellSpecialChars(paramValue);

		// Update args elements with the new parameter value
		this.getTargets("arg")
			.filter((arg) => this.getValue("arg", arg) === paramName)
			.forEach((arg) => {
				arg.textContent = escapedValue || "";
			});

		// Update function links (execute and anchor) with new parameter value
		const functionLinks = [
			...this.getTargets("function-execute"),
			...this.getTargets("function-anchor"),
		] as HTMLAnchorElement[];
		if (functionLinks.length > 0) {
			functionLinks.forEach((functionLink) => {
				const linkAttribute = functionLink.hasAttribute("href")
					? "href"
					: "data-copy-text-value";
				const currentUrl = functionLink.getAttribute(linkAttribute);
				if (!currentUrl) {
					console.warn(
						`No href or data-copy-text-value attribute found for the function link: ${functionLink}.`,
					);
					return;
				}

				const u = new URL(currentUrl, window.location.origin);
				u.searchParams.set(paramName, paramValue);
				functionLink.setAttribute(linkAttribute, u.toString() || "");
			});
		}
	}

	// Update the qeval result
	// If there is no "qeval-result" target, then do nothing.
	private async _updateQEvalResult(): Promise<void> {
		const resultTarget = this.getTarget("qeval-result") as HTMLElement;
		const remoteTarget = this.getTarget("remote") as HTMLElement;

		// If there is no resultTarget or remoteTarget, this is a crossing function.
		if (!(resultTarget && remoteTarget)) return;

		// If there are no args, then show the "(enter param values)" placeholder.
		const argNodes = this.getTargets("arg");
		const haveAllArgs = argNodes.every((arg) => arg.textContent !== "");
		if (!haveAllArgs) {
			resultTarget.textContent = "(enter param values)";
			resultTarget.classList.remove("u-color-danger");
			return;
		}

		// Build the data string for the qeval query.
		const args = argNodes
			.map((arg) => (arg.textContent as string).replace(/\\(.)/g, "$1"))
			.join(",");
		const data = `${this._pkgPath}.${this._funcName}(${args})`;

		// Fetch the qeval result from the remote and update the DOM.
		const result = await this._fetchQEval(remoteTarget.textContent || "", data);
		resultTarget.textContent = result;
		resultTarget.classList.toggle(
			"u-color-danger",
			result.startsWith("Error:"),
		);
	}

	// Fetch the qeval result from the remote
	private async _fetchQEval(remote: string, data: string): Promise<string> {
		try {
			const url = `http://${remote}/abci_query?path=vm%2fqeval&data=${btoa(data)}`;
			const response = await fetch(url);
			if (!response.ok) return "";

			const result = (await response.json()).result.response.ResponseBase;
			return result.Data ? atob(result.Data) : `Error: ${result.Error.value}`;
		} catch {
			return "";
		}
	}

	// DOM ACTIONS
	// update all args (DOM action)
	public updateAllArgs(event: Event): void {
		const target = event.target as HTMLInputElement;
		const paramName = this.getValue("param", target);
		if (!paramName) return;

		// get the current value for the param name
		const paramValue = this._getParamCurrentValue(paramName);
		this._debouncedUpdateAllArgs(paramName, paramValue);
	}

	// update all functions send (DOM action)
	public updateAllFunctionsSend(
		event: Event & { params?: Record<string, unknown> },
	): void {
		const send = (event.params?.send as boolean) || false;
		this.sendValue = send ? this.getValue("send") : null;
		this.getDOMArray("send-code").forEach((sendElement) => {
			sendElement.textContent = this.sendValue || "";
		});
		this._dispatchParamsChanged();
	}
}
