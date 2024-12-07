class Copy {
  private DOM: {
    el: HTMLElement | null;
  };
  private static FEEDBACK_DELAY = 1500;

  private btnClicked: HTMLElement | null = null;
  private btnClickedIcons: HTMLElement[] = [];

  private static SELECTORS = {
    button: "[data-copy-btn]",
    icon: `[data-copy-icon] > use`,
    content: (id: string) => `[data-copy-content="${id}"]`,
  };

  constructor() {
    this.DOM = {
      el: document.querySelector<HTMLElement>("main"),
    };

    if (this.DOM.el) {
      this.init();
    } else {
      console.warn("Copy: Main container not found.");
    }
  }

  private init(): void {
    this.bindEvents();
  }

  private bindEvents(): void {
    this.DOM.el?.addEventListener("click", this.handleClick.bind(this));
  }

  private handleClick(event: Event): void {
    const target = event.target as HTMLElement;
    const button = target.closest<HTMLElement>(Copy.SELECTORS.button);

    if (!button) return;

    this.btnClicked = button;
    this.btnClickedIcons = Array.from(button.querySelectorAll<HTMLElement>(Copy.SELECTORS.icon));

    const contentId = button.getAttribute("data-copy-btn");
    if (!contentId) {
      console.warn("Copy: No content ID found on the button.");
      return;
    }

    const codeBlock = this.DOM.el?.querySelector<HTMLElement>(Copy.SELECTORS.content(contentId));
    if (codeBlock) {
      this.copyToClipboard(codeBlock);
    } else {
      console.warn(`Copy: No content found for ID "${contentId}".`);
    }
  }

  private sanitizeContent(codeBlock: HTMLElement): string {
    const html = codeBlock.innerHTML.replace(/<span class="chroma-ln">.*?<\/span>/g, "");
    const tempDiv = document.createElement("div");
    tempDiv.innerHTML = html;
    return tempDiv.textContent?.trim() || "";
  }

  private toggleIcons(): void {
    this.btnClickedIcons.forEach((icon) => {
      icon.classList.toggle("hidden");
    });
  }

  private showFeedback(): void {
    if (!this.btnClicked) return;

    this.toggleIcons();
    window.setTimeout(() => {
      this.toggleIcons();
    }, Copy.FEEDBACK_DELAY);
  }

  private async copyToClipboard(codeBlock: HTMLElement): Promise<void> {
    const sanitizedText = this.sanitizeContent(codeBlock);

    if (!navigator.clipboard) {
      console.error("Copy: Clipboard API is not supported in this browser.");
      this.showFeedback();
      return;
    }

    try {
      await navigator.clipboard.writeText(sanitizedText);
      console.info("Copy: Text copied successfully.");
      this.showFeedback();
    } catch (err) {
      console.error("Copy: Error while copying text.", err);
      this.showFeedback();
    }
  }
}

export default () => new Copy();
