(() => {
  class Help {
    private DOM: {
      el: HTMLElement | null;
      funcs: HTMLElement[];
      addressInput: HTMLInputElement | null;
      cmdModeSelect: HTMLSelectElement | null;
    };

    private funcList: HelpFunc[];

    constructor() {
      this.DOM = {
        el: document.querySelector("#data"),
        funcs: [],
        addressInput: null,
        cmdModeSelect: null,
      };

      this.funcList = [];

      if (this.DOM.el) this.init();
    }

    private init() {
      const { el } = this.DOM;
      if (!el) return;

      this.DOM.funcs = Array.from(el.querySelectorAll<HTMLElement>("[data-func]"));
      this.DOM.addressInput = el.querySelector<HTMLInputElement>("[data-role='help-input-addr']");
      this.DOM.cmdModeSelect = el.querySelector<HTMLSelectElement>("[data-role='help-select-mode']");

      this.funcList = this.DOM.funcs.map((funcEl) => new HelpFunc(funcEl));

      if (this.DOM.addressInput) this.bindEvents();
    }

    private bindEvents() {
      const { addressInput, cmdModeSelect } = this.DOM;

      addressInput?.addEventListener("input", () => this.funcList.forEach((func) => func.updateAddr(addressInput.value)));

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

    private funcName: null | string = null;

    constructor(el: HTMLElement) {
      this.DOM = {
        el,
        addrs: Array.from(el.querySelectorAll<HTMLElement>("[data-role='help-code-address']")),
        args: Array.from(el.querySelectorAll<HTMLElement>("[data-role='help-code-args']")),
        modes: Array.from(el.querySelectorAll<HTMLElement>("[data-code-mode]")),
      };

      this.funcName = this.DOM.el.dataset.func || "";

      this.bindEvents();
    }

    private bindEvents() {
      this.DOM.el.addEventListener("input", (e) => {
        const target = e.target as HTMLInputElement;
        if (target.dataset.role === "help-param-input") {
          this.updateArg(target.dataset.param || "", target.value);
        }
      });
    }

    public updateArg(paramName: string, paramValue: string) {
      this.DOM.args
        .filter((arg) => arg.dataset.arg === paramName)
        .forEach((arg) => {
          arg.textContent = paramValue.trim() || "";
        });
    }

    public updateAddr(addr: string) {
      this.DOM.addrs.forEach((DOMaddr) => {
        DOMaddr.textContent = addr.trim() || "ADDRESS";
      });
    }

    public updateMode(mode: string) {
      this.DOM.modes.forEach((cmd) => {
        cmd.className = cmd.dataset.codeMode === mode ? "inline" : "hidden";
        cmd.dataset.copyContent = cmd.dataset.codeMode === mode ? `help-cmd-${this.funcName}` : "";
      });
    }
  }

  document.addEventListener("DOMContentLoaded", () => new Help());
})();
