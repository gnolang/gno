import {
	addressSlotText,
	argSlotText,
	BaseController,
	debounce,
	evalArgs,
	paramValue,
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

		// Served hidden: Eval only works from here on.
		const evalButton = this.getTarget("eval-button");
		if (evalButton) evalButton.hidden = false;

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
			addressElement.textContent = addressSlotText(address);
		});
	}

	// get current value for a param name (handles checkbox multiple values)
	private _getParamCurrentValue(paramName: string): string {
		return this._inputsValue(
			this._paramInputs.filter(
				(inp) => this.getValue("param", inp) === paramName,
			),
		);
	}

	// the value of one param's inputs: several for a radio or a checkbox
	private _inputsValue(inputs: HTMLInputElement[]): string {
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

		// Other: the value the live result and Eval evaluate too (evalArgs).
		return paramValue(firstInput.value, this.getValue("type", firstInput));
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

			const value = this._getParamCurrentValue(paramName);
			this._params[paramName] = value;
			if (value) this._updateArgInURLs(paramName, value);

			processed.add(paramName);
		});
		this._fillArgSlots();
	}

	// debounced update all args and update the qeval result
	private _debouncedUpdateAllArgs = debounce((paramName: string) => {
		this._fillArgSlots();
		this._debouncedQEval();
		if (!paramName) return;
		this._updateArgInURLs(paramName, this._getParamCurrentValue(paramName));
		this._dispatchParamsChanged();
	}, 50);

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

	// Fills the printed commands' arg slots. A slot and its param's inputs
	// share a key: the param's position on $help, where names repeat or are
	// missing ("_"); its name in a form, whose inputs carry no position.
	private _fillArgSlots(): void {
		const inputs = new Map<string, HTMLInputElement[]>();
		for (const input of this._paramInputs) {
			const key = this.getValue("arg", input) || this.getValue("param", input);
			const group = inputs.get(key);
			if (group) group.push(input);
			else inputs.set(key, [input]);
		}
		this.getTargets("arg").forEach((slot) => {
			slot.textContent = argSlotText(
				this._inputsValue(inputs.get(this.getValue("arg", slot)) ?? []),
				this.getValue("type", slot),
			);
		});
	}

	// Update function execute (form) and anchor (copy button) with new parameter
	private _updateArgInURLs(paramName: string, value: string): void {
		const executeForm = this.getTarget("function-execute") as HTMLFormElement;
		const anchorButton = this.getTarget("function-anchor") as HTMLButtonElement;
		if (!executeForm && !anchorButton) return;

		const baseUrl =
			executeForm?.getAttribute("action") ||
			anchorButton?.getAttribute("data-copy-text-value");
		if (!baseUrl) return;

		const parts = baseUrl.split("&");
		const encodedValue = encodeURIComponent(value);
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

		// get the current value for the param name and update params tree
		if (paramName) {
			this._params[paramName] = this._getParamCurrentValue(paramName);
		}
		this._debouncedUpdateAllArgs(paramName);
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
