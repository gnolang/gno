(() => {
  class Copy {
    private DOM: {
      el: HTMLElement | null;
      btnClicked: HTMLElement | null;
    };

    constructor() {
      this.DOM = {
        el: document.querySelector("main"),
        btnClicked: null,
      };

      if (this.DOM.el) this.init();
    }

    private init() {
      this.bindEvents();
    }

    private bindEvents() {
      this.DOM.el?.addEventListener("click", this.handleClick.bind(this));
    }

    private handleClick(event: Event) {
      const target = event.target as HTMLElement;
      const button = target.closest<HTMLElement>("[data-copy-btn]");

      if (!button) return;

      this.DOM.btnClicked = button;
      const contentId = button.getAttribute("data-copy-btn");
      if (!contentId) return;

      const codeBlock = this.DOM.el?.querySelector<HTMLElement>(`[data-copy-content="${contentId}"]`);
      if (codeBlock) this.copyToClipboard(codeBlock);
    }

    private sanitizeContent(codeBlock: HTMLElement): string {
      const clonedBlock = codeBlock.cloneNode(true) as HTMLElement;

      clonedBlock.querySelectorAll(".chroma-ln").forEach((lineNumber) => lineNumber.remove());

      return clonedBlock.textContent?.trim() || "";
    }

    private showFeedback(success: boolean) {
      if (!this.DOM.btnClicked) return;

      const feedbackClass = success ? "text-green-600" : "";

      this.DOM.btnClicked.classList.add(feedbackClass);
      window.setTimeout(() => {
        this.DOM.btnClicked?.classList.remove(feedbackClass);
      }, 1500);
    }

    async copyToClipboard(codeBlock: HTMLElement) {
      const sanitizedText = this.sanitizeContent(codeBlock);
      try {
        await navigator.clipboard.writeText(sanitizedText);
        this.showFeedback(true);
      } catch (err) {
        console.error("Copy error: ", err);
        this.showFeedback(false);
      }
    }
  }

  document.addEventListener("DOMContentLoaded", () => new Copy());
})();
