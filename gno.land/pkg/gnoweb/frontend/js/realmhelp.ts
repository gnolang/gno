class Help {
  private DOM: {
    el: HTMLElement | null;
    funcs: HTMLElement[];
    addressInput: HTMLInputElement | null;
    cmdModeSelect: HTMLSelectElement | null;
  };

  private funcList: HelpFunc[];

  private static SELECTORS = {
    container: "#help",
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

    this.DOM.funcs = Array.from(el.querySelectorAll<HTMLElement>(Help.SELECTORS.func));
    this.DOM.addressInput = el.querySelector<HTMLInputElement>(Help.SELECTORS.addressInput);
    this.DOM.cmdModeSelect = el.querySelector<HTMLSelectElement>(Help.SELECTORS.cmdModeSelect);

    console.log(this.DOM);
    this.funcList = this.DOM.funcs.map((funcEl) => new HelpFunc(funcEl));

    this.bindEvents();
  }

  private bindEvents(): void {
    const { addressInput, cmdModeSelect } = this.DOM;

    addressInput?.addEventListener("input", () => {
      this.funcList.forEach((func) => func.updateAddr(addressInput.value));
    });

    cmdModeSelect?.addEventListener("change", (e) => {
      const target = e.target as HTMLSelectElement;
      this.funcList.forEach((func) => func.updateMode(target.value));
    });
  }
}

class HelpFunc {
  private DOM: {
    el: HTMLElement;
    addrs: HTMLElement[];
    args: HTMLElement[];
    modes: HTMLElement[];
  };

  private funcName: string | null;

  private static SELECTORS = {
    address: "[data-role='help-code-address']",
    args: "[data-role='help-code-args']",
    mode: "[data-code-mode]",
    paramInput: "[data-role='help-param-input']",
  };

  constructor(el: HTMLElement) {
    this.DOM = {
      el,
      addrs: Array.from(el.querySelectorAll<HTMLElement>(HelpFunc.SELECTORS.address)),
      args: Array.from(el.querySelectorAll<HTMLElement>(HelpFunc.SELECTORS.args)),
      modes: Array.from(el.querySelectorAll<HTMLElement>(HelpFunc.SELECTORS.mode)),
    };

    this.funcName = el.dataset.func || null;

    this.bindEvents();
  }

  private bindEvents(): void {
    this.DOM.el.addEventListener("input", (e) => {
      const target = e.target as HTMLInputElement;
      if (target.dataset.role === "help-param-input") {
        this.updateArg(target.dataset.param || "", target.value);
      }
    });
  }

  public updateArg(paramName: string, paramValue: string): void {
    this.DOM.args
      .filter((arg) => arg.dataset.arg === paramName)
      .forEach((arg) => {
        arg.textContent = paramValue.trim() || "";
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
      cmd.className = isVisible ? "inline" : "hidden";
      cmd.dataset.copyContent = isVisible ? `help-cmd-${this.funcName}` : "";
    });
  }
}

export default () => new Help();
