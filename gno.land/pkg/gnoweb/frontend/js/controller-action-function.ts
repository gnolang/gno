import {
	BaseController,
	debounce,
	escapeShellSpecialChars,
	evalArgs,
	gnoLiteral,
	postEval,
} from "./controller.js";
import type { ActionMode } from "./controller-action-header.js";

export class ActionFunctionController extends BaseController {
	protected sendValue: string | null = null;
	declare _funcName: string | null;
	declare _pkgPath: string | null;
	declare _paramInputsCache: HTMLInputElement[];
	declare _params: Record<string, string>;
	// Set in connect(), which runs before class-field initialisers.
	declare _qevalSeq: number;
	// The live result's request in flight, aborted when a newer one starts.
	declare _qevalAbort: AbortController | null;

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

		this._qevalSeq = 0;
		this._qevalAbort = null;
		this._funcName = this.getValue("name") || null;
		this._pkgPath = this.getValue("pkgpath") || null;

		this._initializeArgs();
		this._listenForEvents();

		// Some functions may have no params, or all params already have values
		this._updateQEvalResult();

		// Dispatch initial params state for wallet integration
		this._dispatchParamsChanged();
	}

	protected disconnect(): void {
		this._qevalAbort?.abort();
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
				.filter((inp) => inp.checked)
				.map((inp) => inp.value.trim())
				.join(",");
		}

		// Radio: find checked one
		if (firstInput.type === "radio") {
			const checked = inputs.find((inp) => inp.checked);

			return checked?.value.trim() || "";
		}

		// Other: return value
		return firstInput.value.trim();
	}

	// initialize the args
	private _initializeArgs(): void {
		// multiple values (radio or checkbox) to be initialized only once
		const processed = new Set<string>();

		// Build params tree and initialize the args
		this._params = {};
		this._paramInputs.forEach((input) => {
			const paramName = this.getValue("param", input) || "";

			if (!paramName || processed.has(paramName)) return;

			const paramValue = this._getParamCurrentValue(paramName);
			this._params[paramName] = paramValue;
			if (paramValue) this._updateArgInDOM(paramName, paramValue);

			processed.add(paramName);
		});
	}

	// debounced update all args and update the qeval result
	private _debouncedUpdateAllArgs = debounce(
		(paramName: string, paramValue: string) => {
			if (paramName) {
				this._updateArgInDOM(paramName, paramValue);
				this._debouncedQEval();
				this._dispatchParamsChanged();
			}
		},
		50,
	);

	// The live result queries the rate-limited /_/api/eval: wait for a pause
	// in typing instead of sending one request per keystroke.
	private _debouncedQEval = debounce(() => this._updateQEvalResult(), 400);

	// Dispatch params:changed event uppon parameter updates.
	private _dispatchParamsChanged(): void {
		if (!this._funcName || !this._pkgPath) return;

		this.dispatch("params:changed", {
			pkgPath: this._pkgPath,
			funcName: this._funcName,
			params: { ...this._params },
			send: this.sendValue || undefined,
		});
	}

	// push args in DOM (in func code)
	private _updateArgInDOM(paramName: string, paramValue: string): void {
		// Update args elements with the new parameter value. A qeval arg
		// carries its param type, and is a Gno literal for it; a maketx call
		// arg has none and stays the raw value gnokey parses.
		this.getTargets("arg")
			.filter((arg) => this.getValue("arg", arg) === paramName)
			.forEach((arg) => {
				const type = this.getValue("type", arg);
				arg.textContent = paramValue
					? escapeShellSpecialChars(gnoLiteral(paramValue, type))
					: "";
			});

		// Update function execute (form) and anchor (copy button) with new parameter
		const executeForm = this.getTarget("function-execute") as HTMLFormElement;
		const anchorButton = this.getTarget("function-anchor") as HTMLButtonElement;
		if (!executeForm && !anchorButton) return;

		const baseUrl =
			executeForm?.getAttribute("action") ||
			anchorButton?.getAttribute("data-copy-text-value");
		if (!baseUrl) return;

		const parts = baseUrl.split("&");
		const encodedValue = encodeURIComponent(paramValue);
		const index = parts.findIndex(
			(part, i) => i > 0 && part.startsWith(`${paramName}=`),
		);
		if (index !== -1) {
			parts[index] = `${paramName}=${encodedValue}`;
		} else {
			parts.push(`${paramName}=${encodedValue}`);
		}

		const updatedUrl = parts.join("&");

		executeForm?.setAttribute("action", updatedUrl);
		anchorButton?.setAttribute("data-copy-text-value", updatedUrl);
	}

	// Update the qeval result
	// If there is no "qeval-result" target, then do nothing.
	// Placeholder texts and the error class are all read from data-* attributes
	// on the target so the contract is shared with analytics.ts.
	private async _updateQEvalResult(): Promise<void> {
		// Any newer update, even one that only shows the placeholder, voids
		// the request still in flight: abort it, and drop its answer should
		// it land anyway.
		const seq = ++this._qevalSeq;
		this._qevalAbort?.abort();
		this._qevalAbort = null;
		const resultTarget = this.getTarget("qeval-result") as HTMLElement;

		// Crossing functions have no qeval-result target.
		if (!resultTarget) return;

		const placeholder = resultTarget.dataset.qevalPlaceholder ?? "";
		const unavailable = resultTarget.dataset.qevalUnavailable ?? "";
		const errorClass = resultTarget.dataset.qevalErrorClass ?? "u-color-danger";

		// Until every param has a value, show the placeholder.
		const { args, full } = evalArgs(this._paramInputs);
		if (!full) {
			resultTarget.textContent = placeholder;
			resultTarget.classList.remove(errorClass);
			return;
		}

		// The same expression as the Eval button's: one Gno literal per param.
		const expression = `${this._funcName}(${args.join(", ")})`;

		const abort = new AbortController();
		this._qevalAbort = abort;
		const res = await postEval(this._pkgPath ?? "", expression, abort.signal);
		if (seq !== this._qevalSeq) return;
		this._qevalAbort = null;
		// A rate limit or an outage says nothing about the function: show the
		// neutral text analytics ignores, not an error.
		resultTarget.textContent = res.unavailable ? unavailable : res.text;
		resultTarget.classList.toggle(errorClass, res.isError);
	}

	// DOM ACTIONS

	// Dispatch eval:expression event (DOM action), built from the params
	// typed in the card. With every param filled (or none), the evaluator
	// runs it; otherwise it selects the arguments for the reader to finish.
	public triggerEval(): void {
		const { args, full } = evalArgs(this._paramInputs);
		this.dispatch("eval:expression", {
			expression: `${this._funcName || ""}(${args.join(", ")})`,
			autoEval: full,
		});
	}

	// update all args (DOM action)
	public updateAllArgs(event: Event): void {
		const target = event.target as HTMLInputElement;
		const paramName = this.getValue("param", target);
		if (!paramName) return;

		// get the current value for the param name and update params tree
		const paramValue = this._getParamCurrentValue(paramName);
		this._params[paramName] = paramValue;
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
