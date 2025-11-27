import {
	BaseController,
	debounce,
	escapeShellSpecialChars,
} from "./controller.js";
import type { ActionMode } from "./controller-action-header.js";

export class ActionFunctionController extends BaseController {
	protected sendValue: string | null = null;
	declare _funcName: string | null;

	protected connect(): void {
		this.initializeDOM({
			"send-code": this.getTargets("send-code"),
		});

		this._funcName = this.getValue("name") || null;

		this._initializeArgs();
		this._listenForEvents();
	}

	// listen for events from action-header controller
	private _listenForEvents(): void {
		// Listen for mode changes from action-header controller
		this.on("mode:changed", (event: Event) => {
			const customEvent = event as CustomEvent;
			const mode: ActionMode = customEvent.detail.mode;
			console.log("mode:changed", mode);
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
			console.log("modeElement in function", modeElement);
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

	// initialize the args
	private _initializeArgs(): void {
		this.getTargets("param-input").forEach((paramInput) => {
			const { paramName, paramValue } = this._sanitizeArgsInput(
				paramInput as HTMLInputElement,
			);
			if (paramName) this._pushArgsInDOM(paramName, paramValue);
		});
	}

	// debounced update all args
	private _debouncedUpdateAllArgs = debounce(
		(paramName: string, paramValue: string) => {
			if (paramName) this._pushArgsInDOM(paramName, paramValue);
		},
		50,
	);

	// push args in DOM (in func code)
	private _pushArgsInDOM(paramName: string, paramValue: string): void {
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
					console.warn(`No href attribute found for function link`);
					return;
				}

				const u = new URL(currentUrl, window.location.origin);
				u.searchParams.set(paramName, paramValue);
				functionLink.setAttribute(linkAttribute, u.toString() || "");
			});
		}
	}

	// DOM ACTIONS
	// update all args (DOM action)
	public updateAllArgs(event: Event): void {
		const target = event.target as HTMLInputElement;
		const { paramName, paramValue } = this._sanitizeArgsInput(target);

		if (paramName) this._debouncedUpdateAllArgs(paramName, paramValue);
	}

	// update all functions send (DOM action)
	public updateAllFunctionsSend(
		event: Event & { params?: Record<string, unknown> },
	): void {
		const send = (event.params?.send as boolean) || false;
		this.getDOMArray("send-code").forEach((sendElement) => {
			sendElement.textContent = send ? this.getValue("send") : "";
		});
	}
}
