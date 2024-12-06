(() => {
  class Help {
    constructor() {
      this.DOM = {
        el: document.querySelector("#data"),
        funcs: [],
        addressInput: null,
        cmdModeSelect: null,
      };

      this.funcList = [];

      if (this.DOM.el) this.#init();
    }

    #init() {
      const { el } = this.DOM;
      this.DOM.funcs = [...el.querySelectorAll("[data-func]")];
      this.DOM.addressInput = el.querySelector("[data-role='help-input-addr']");
      this.DOM.cmdModeSelect = el.querySelector("[data-role='help-select-mode']");

      this.funcList = this.DOM.funcs.map((funcEl) => new HelpFunc(funcEl));

      if (this.DOM.addressInput) this.#bindEvents();
    }

    #bindEvents() {
      const { addressInput, cmdModeSelect } = this.DOM;

      addressInput.addEventListener("input", () => this.funcList.forEach((func) => func.updateAddr(addressInput.value)));
      cmdModeSelect?.addEventListener("change", (e) => this.funcList.forEach((func) => func.updateMode(e.target.value)));
    }
  }

  class HelpFunc {
    constructor(el) {
      this.DOM = {
        el,
        addrs: [...el.querySelectorAll("[data-role='help-code-address']")],
        args: [...el.querySelectorAll("[data-role='help-code-args']")],
        modes: [...el.querySelectorAll("[data-code-mode]")],
      };

      this.#bindEvents();
    }

    #bindEvents() {
      this.DOM.el.addEventListener("input", (e) => {
        if (e.target.dataset.role === "help-param-input") {
          this.updateArg(e.target.dataset.param, e.target.value);
        }
      });
    }

    updateArg(paramName, paramValue) {
      this.DOM.args.filter((arg) => arg.dataset.arg === paramName).forEach((arg) => (arg.textContent = paramValue.trim() || ""));
    }

    updateAddr(addr) {
      this.DOM.addrs.forEach((DOMaddr) => (DOMaddr.textContent = addr.trim() || "ADDRESS"));
    }

    updateMode(mode) {
      this.DOM.modes.forEach((cmd) => {
        cmd.className = cmd.dataset.codeMode === mode ? "inline" : "hidden";
      });
    }
  }

  document.addEventListener("DOMContentLoaded", () => new Help());
})();
