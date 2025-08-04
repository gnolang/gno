import { debounce, escapeShellSpecialChars } from "./utils";

class Help {
	private DOM: {
		el: HTMLElement | null;
		funcs: HTMLElement[];
		addressInput: HTMLInputElement | null;
		cmdModeSelect: HTMLSelectElement | null;
	};

	private funcList: HelpFunc[];

	private static SELECTORS = {
		container: ".js-help-view",
		func: "[data-func]",
		addressInput: "[data-role='help-input-addr']",
		cmdModeSelect: "[data-role='help-select-mode']",
	};

	constructor() {
		this.DOM = {
			el: document.querySelector<HTMLElement>(Help.SELECTORS.container),
			funcs: [],
			addressInput: null,
			cmdModeSelect: null,
		};

		this.funcList = [];

		if (this.DOM.el) {
			this.init();
		} else {
			console.warn("Help: Main container not found.");
		}
	}

	private init(): void {
		const { el } = this.DOM;
		if (!el) return;

		this.DOM.funcs = Array.from(
			el.querySelectorAll<HTMLElement>(Help.SELECTORS.func),
		);
		this.DOM.addressInput = el.querySelector<HTMLInputElement>(
			Help.SELECTORS.addressInput,
		);
		this.DOM.cmdModeSelect = el.querySelector<HTMLSelectElement>(
			Help.SELECTORS.cmdModeSelect,
		);

		this.funcList = this.DOM.funcs.map((funcEl) => new HelpFunc(funcEl));

		this.restoreAddress();
		this.restoreMode();
		this.bindEvents();
	}

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

	private restoreAddress(): void {
		this.restoreValue("helpAddressInput", this.DOM.addressInput, (value) =>
			this.funcList.forEach((func) => func.updateAddr(value)),
		);
	}

	private restoreMode(): void {
		this.restoreValue("helpCmdMode", this.DOM.cmdModeSelect, (value) =>
			this.funcList.forEach((func) => func.updateMode(value)),
		);
	}

	private bindEvents(): void {
		const { addressInput, cmdModeSelect } = this.DOM;

		const debouncedUpdate = debounce((addressInput: HTMLInputElement) => {
			const address = addressInput.value;
			localStorage.setItem("helpAddressInput", address);
			this.funcList.forEach((func) => func.updateAddr(address));
		}, 50);
		addressInput?.addEventListener("input", () =>
			debouncedUpdate(addressInput),
		);

		cmdModeSelect?.addEventListener("change", (e) => {
			const target = e.target as HTMLSelectElement;
			const mode = target.value;
			localStorage.setItem("helpCmdMode", mode);
			this.funcList.forEach((func) => func.updateMode(mode));
		});
	}
}

class HelpFunc {
	private DOM: {
		el: HTMLElement;
		addrs: HTMLElement[];
		args: HTMLElement[];
		modes: HTMLElement[];
		paramInputs: HTMLInputElement[];
		functionLink: HTMLAnchorElement | null;
		send: HTMLElement[];
		sendInput: HTMLInputElement | null;
	};

	private funcName: string | null;

	private static SELECTORS = {
		address: "[data-role='help-code-address']",
		args: "[data-role='help-code-args']",
		mode: "[data-code-mode]",
		paramInput: "[data-role='help-param-input']",
		functionLink: "[data-role='help-function-link']",
		send: "[data-role='help-code-send']",
		sendInput: "[data-role='help-send-input']",
	};

	constructor(el: HTMLElement) {
		this.DOM = {
			el,
			addrs: Array.from(
				el.querySelectorAll<HTMLElement>(HelpFunc.SELECTORS.address),
			),
			args: Array.from(
				el.querySelectorAll<HTMLElement>(HelpFunc.SELECTORS.args),
			),
			modes: Array.from(
				el.querySelectorAll<HTMLElement>(HelpFunc.SELECTORS.mode),
			),
			paramInputs: Array.from(
				el.querySelectorAll<HTMLInputElement>(HelpFunc.SELECTORS.paramInput),
			),
			functionLink: el.querySelector<HTMLAnchorElement>(
				HelpFunc.SELECTORS.functionLink,
			),
			send: Array.from(
				el.querySelectorAll<HTMLElement>(HelpFunc.SELECTORS.send),
			),
			sendInput: el.querySelector<HTMLInputElement>(
				HelpFunc.SELECTORS.sendInput,
			),
		};

		this.funcName = el.dataset.func || null;

		this.initializeArgs();
		this.bindEvents();
	}

	private static sanitizeArgsInput(input: HTMLInputElement) {
		const paramName = input.dataset.param || "";
		const paramValue = input.value.trim();

		if (!paramName) {
			console.warn("sanitizeArgsInput: param is missing in arg input dataset.");
		}

		return { paramName, paramValue };
	}

	private bindEvents(): void {
		const debouncedUpdate = debounce(
			(paramName: string, paramValue: string) => {
				if (paramName) this.updateArg(paramName, paramValue);
			},
			50,
		);

		this.DOM.el.addEventListener("input", (e) => {
			const target = e.target as HTMLInputElement;
			if (target.dataset.role === "help-param-input") {
				const { paramName, paramValue } = HelpFunc.sanitizeArgsInput(target);
				debouncedUpdate(paramName, paramValue);
			}
		});

		this.DOM.el.addEventListener("change", (e) => {
			const target = e.target as HTMLInputElement;
			if (target.dataset.role === "help-send-input") {
				this.updateSend(target.checked);
			}
		});
	}

	private initializeArgs(): void {
		this.DOM.paramInputs.forEach((input) => {
			const { paramName, paramValue } = HelpFunc.sanitizeArgsInput(input);
			if (paramName) this.updateArg(paramName, paramValue);
		});
	}

	public updateArg(paramName: string, paramValue: string): void {
		const escapedValue = escapeShellSpecialChars(paramValue);
		this.DOM.args
			.filter((arg) => arg.dataset.arg === paramName)
			.forEach((arg) => {
				arg.textContent = escapedValue || "";
			});

		// Update function link with new parameter value
		if (this.DOM.functionLink) {
			const currentUrl = this.DOM.functionLink.getAttribute("href");
			if (!currentUrl) {
				console.warn(`No href attribute found for function ${this.funcName}`);
				return;
			}

			if (paramValue) {
				const paramRegex = new RegExp(`&${paramName}=[^&]*`);
				const newParam = `&${paramName}=${encodeURIComponent(paramValue)}`;

				const newUrl = paramRegex.test(currentUrl)
					? currentUrl.replace(paramRegex, newParam)
					: `${currentUrl}${newParam}`;

				this.DOM.functionLink.setAttribute("href", newUrl);
			}
		}
	}

	public updateSend(sendValue: boolean): void {
		const sendAmount = sendValue
			? (this.DOM.sendInput?.dataset.send ?? "")
			: "";
		this.DOM.send.forEach((send) => {
			send.textContent = sendAmount;
		});
	}

	public updateAddr(addr: string): void {
		this.DOM.addrs.forEach((DOMaddr) => {
			DOMaddr.textContent = addr.trim() || "ADDRESS";
		});
	}

	public updateMode(mode: string): void {
		this.DOM.modes.forEach((cmd) => {
			const isVisible = cmd.dataset.codeMode === mode;
			cmd.classList.toggle("inline", isVisible);
			cmd.classList.toggle("hidden", !isVisible);
			cmd.dataset.copyContent = isVisible ? `help-cmd-${this.funcName}` : "";
		});
	}
}

export default () => new Help();
