import { BaseController } from "./controller.js";

export class CopyController extends BaseController {
	private static FEEDBACK_DELAY = 750;
	private isAnimationRunning = false;
	private btnClicked: HTMLElement | null = null;

	protected connect(): void {}

	// sanitize content (remove comments)
	private _sanitizeContent(
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

	// toggle icons (show/hide)
	private _toggleIcons(icons: HTMLElement[]): void {
		icons.forEach((icon) => {
			icon.classList.toggle("u-hidden");
		});
	}

	// show feedback (animation)
	private _showFeedback(icons: HTMLElement[]): void {
		if (!this.btnClicked || this.isAnimationRunning === true) return;

		this.isAnimationRunning = true;
		this._toggleIcons(icons);
		window.setTimeout(() => {
			this._toggleIcons(icons);
			this.isAnimationRunning = false;
		}, CopyController.FEEDBACK_DELAY);
	}

	// utility method to write to clipboard
	private async _writeToClipboard(
		text: string,
		icons: HTMLElement[],
	): Promise<void> {
		if (!navigator.clipboard) {
			console.error("Copy: Clipboard API is not supported in this browser.");
			this._showFeedback(icons);
			return;
		}

		try {
			await navigator.clipboard.writeText(text);
			this._showFeedback(icons);
		} catch (err) {
			console.error("Copy: Error while copying text.", err);
			this._showFeedback(icons);
		}
	}

	// copy text to clipboard
	private async _copyTextToClipboard(
		text: string,
		icons: HTMLElement[],
	): Promise<void> {
		const cleaned = text.trim();
		await this._writeToClipboard(cleaned, icons);
	}

	// copy code block to clipboard
	private async _copyToClipboard(
		codeBlock: HTMLElement,
		icons: HTMLElement[],
		removeComments: boolean = false,
	): Promise<void> {
		const sanitizedText = this._sanitizeContent(codeBlock, removeComments);
		await this._writeToClipboard(sanitizedText, icons);
	}

	// DOM ACTIONS
	// handle click event (DOM action)
	public copy(event: Event): void {
		this.btnClicked = event.currentTarget as HTMLElement;
		if (this.isAnimationRunning) return;

		const icons = this.getTargets("icon");
		const btnClickedIcons = Array.from(icons);

		// Handle data-copy-text (direct text)
		if (this.getValue("text")) {
			this._copyTextToClipboard(this.getValue("text"), btnClickedIcons);
			return;
		}

		// Handle data-copy-remote (remote content in DOM)
		if (this.getValue("remote")) {
			const target = this.getGlobalTarget(this.getValue("remote"));
			if (!target) {
				console.warn(`Copy: No target found for "${this.getValue("remote")}".`);
				return;
			}
			const clean = this.hasValue("clean");
			this._copyToClipboard(target, btnClickedIcons, clean);
			return;
		}

		console.warn("Copy: No content to copy found on the button.");
	}
}
