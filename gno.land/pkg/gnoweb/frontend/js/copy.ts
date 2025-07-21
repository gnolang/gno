class Copy {
	private DOM: {
		el: HTMLElement | null;
	};
	private static FEEDBACK_DELAY = 750;

	private btnClicked: HTMLElement | null = null;
	private btnClickedIcons: HTMLElement[] = [];
	private isAnimationRunning: boolean = false;

	private static SELECTORS = {
		button: ".js-copy-btn",
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
		const button = (event.target as HTMLElement).closest<HTMLElement>(
			Copy.SELECTORS.button,
		);
		if (!button) return;

		this.btnClicked = button;
		this.btnClickedIcons = Array.from(
			button.querySelectorAll<HTMLElement>(Copy.SELECTORS.icon),
		);

		// Handle data-copy-txt (direct text)
		const contentTxt = button.getAttribute("data-copy-txt");
		if (contentTxt) {
			this.copyTextToClipboard(contentTxt, this.btnClickedIcons);
			return;
		}

		// Handle data-copy-target (legacy selector)
		const contentSrc = button.getAttribute("data-copy-target");
		if (contentSrc) {
			const codeBlock = this.DOM.el?.querySelector<HTMLElement>(
				Copy.SELECTORS.content(contentSrc),
			);
			if (!codeBlock) {
				console.warn(`Copy: No content found for source "${contentSrc}".`);
				return;
			}
			this.copyToClipboard(codeBlock, this.btnClickedIcons, false);
			return;
		}

		// Handle data-copy-btn (new selector with optional comment removal)
		const contentId = button.getAttribute("data-copy-btn");
		if (contentId) {
			const codeBlock = this.DOM.el?.querySelector<HTMLElement>(
				Copy.SELECTORS.content(contentId),
			);
			if (codeBlock) {
				const removeComments = button.hasAttribute("data-copy-remove-comments");
				this.copyToClipboard(codeBlock, this.btnClickedIcons, removeComments);
			} else {
				console.warn(`Copy: No content found for ID "${contentId}".`);
			}
			return;
		}

		console.warn("Copy: No content to copy found on the button.");
	}

	private sanitizeContent(
		codeBlock: HTMLElement,
		removeComments: boolean = false,
	): string {
		const html = codeBlock.innerHTML.replace(
			/<span[^>]*class="chroma-ln"[^>]*>[\s\S]*?<\/span>/g,
			"",
		);

		const tempDiv = document.createElement("div");
		tempDiv.innerHTML = html;

		let text = tempDiv.textContent?.trim() || "";

		if (removeComments) {
			text = text
				.split("\n")
				.filter((line) => {
					const trimmed = line.trim();
					return trimmed && !trimmed.match(/^[#/*]/);
				})
				.join("\n");
		}

		return text;
	}

	private toggleIcons(icons: HTMLElement[]): void {
		icons.forEach((icon) => {
			icon.classList.toggle("hidden");
		});
	}

	private showFeedback(icons: HTMLElement[]): void {
		if (!this.btnClicked || this.isAnimationRunning === true) return;

		this.isAnimationRunning = true;
		this.toggleIcons(icons);
		window.setTimeout(() => {
			this.toggleIcons(icons);
			this.isAnimationRunning = false;
		}, Copy.FEEDBACK_DELAY);
	}

	private async copyTextToClipboard(
		text: string,
		icons: HTMLElement[],
	): Promise<void> {
		if (!navigator.clipboard) {
			console.error("Copy: Clipboard API is not supported in this browser.");
			this.showFeedback(icons);
			return;
		}

		try {
			await navigator.clipboard.writeText(text.trim());
			this.showFeedback(icons);
		} catch (err) {
			console.error("Copy: Error while copying text.", err);
			this.showFeedback(icons);
		}
	}

	private async copyToClipboard(
		codeBlock: HTMLElement,
		icons: HTMLElement[],
		removeComments: boolean = false,
	): Promise<void> {
		const sanitizedText = this.sanitizeContent(codeBlock, removeComments);

		if (!navigator.clipboard) {
			console.error("Copy: Clipboard API is not supported in this browser.");
			this.showFeedback(icons);
			return;
		}

		try {
			await navigator.clipboard.writeText(sanitizedText);
			this.showFeedback(icons);
		} catch (err) {
			console.error("Copy: Error while copying text.", err);
			this.showFeedback(icons);
		}
	}
}

export default () => new Copy();
