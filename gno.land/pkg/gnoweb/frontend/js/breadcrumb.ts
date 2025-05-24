class Breadcrumb {
  private DOM: {
    inputs: HTMLInputElement[];
  };

  private static SELECTORS = {
    breadcrumb: "[data-role='header-breadcrumb-search']",
  };

  constructor(el: HTMLElement) {
    this.DOM = {
      inputs: Array.from(
        el.querySelectorAll<HTMLInputElement>(`${Breadcrumb.SELECTORS.breadcrumb} input[type="text"]`)
      ),
    };

    if (this.DOM.inputs.length) {
        this.bindEvents();
    }
  }

  private bindEvents(): void {
    console.log(this.DOM.inputs);
    // Select the input value when it is focused
    this.DOM.inputs.forEach(input => {
      input.addEventListener("focus", () => {
        input.select();
      });
    });
  }
}

export default (el: HTMLElement) => new Breadcrumb(el); 