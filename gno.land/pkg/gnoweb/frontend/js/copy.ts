(() => {
  class Copy {
    private DOM: {
      el: HTMLElement | null;
    };
    btnClicked: HTMLElement | null = null;
    btnClickedIcons: HTMLElement[] = [];

    constructor() {
      this.DOM = {
        el: document.querySelector("main"),
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

      this.btnClicked = button;
      this.btnClickedIcons = Array.from(button.querySelectorAll<HTMLElement>(`[data-copy-icon] > use`));

      const contentId = button.getAttribute("data-copy-btn");
      if (!contentId) return;

      const codeBlock = this.DOM.el?.querySelector<HTMLElement>(`[data-copy-content="${contentId}"]`);
      if (codeBlock) this.copyToClipboard(codeBlock);
    }

    private sanitizeContent(codeBlock: HTMLElement): string {
      const html = codeBlock.innerHTML.replace(/<span class="chroma-ln">.*?<\/span>/g, "");
      const tempDiv = document.createElement("div");
      tempDiv.innerHTML = html;
      return tempDiv.textContent?.trim() || "";
    }

    private toggleIcons() {
      this.btnClickedIcons.forEach((icon) => {
        icon.classList.toggle("hidden");
      });
    }

    private showFeedback() {
      if (!this.btnClicked) return;

      this.toggleIcons();
      window.setTimeout(() => {
        this.toggleIcons();
      }, 1500);
    }

    async copyToClipboard(codeBlock: HTMLElement) {
      const sanitizedText = this.sanitizeContent(codeBlock);
      try {
        await navigator.clipboard.writeText(sanitizedText);
        this.showFeedback();
      } catch (err) {
        console.error("Copy error: ", err);
        this.showFeedback();
      }
    }
  }

  document.addEventListener("DOMContentLoaded", () => new Copy());
})();
