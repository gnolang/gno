class Breadcrumb {
  private DOM: {
    inputs: HTMLInputElement[];
  };

  constructor(el: HTMLElement) {
    this.DOM = {
      inputs: Array.from(
        el.querySelectorAll<HTMLInputElement>(`input[type="text"]`)
      ),
    };

    if (this.DOM.inputs.length) {
        this.bindEvents();
    }
  }

  private bindEvents(): void {
    this.DOM.inputs.forEach(input => {
      input.addEventListener("focus", () => input.select());
    });
  }
}

export default (el: HTMLElement) => new Breadcrumb(el); 