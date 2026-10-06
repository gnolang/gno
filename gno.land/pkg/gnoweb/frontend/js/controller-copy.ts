import { BaseController } from "./controller.js";

export class CopyController extends BaseController {
	private static FEEDBACK_DELAY = 750;
	private static FAILURE_DELAY = 2000;
	private static COPIED_TEXT = "Copied";
	private static FAILED_TEXT = "Copy failed";
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

	// announce the result to assistive tech, through an optional
	// role="status" target (data-copy-target="status") inside the controller.
	// Its text is reset first so the same message is announced again.
	private _announce(message: string): void {
		const status = this.getTarget("status");
		if (!status) return;
		status.textContent = "";
		window.setTimeout(() => {
			status.textContent = message;
		}, 50);
	}

	// show feedback (animation): the check icon, on success only
	private _showFeedback(icons: HTMLElement[]): void {
		if (!this.btnClicked || this.isAnimationRunning === true) return;

		this.isAnimationRunning = true;
		this._toggleIcons(icons);
		this._announce(this.getValue("copied") || CopyController.COPIED_TEXT);
		window.setTimeout(() => {
			this._toggleIcons(icons);
			this.isAnimationRunning = false;
		}, CopyController.FEEDBACK_DELAY);
	}

	// show a failure: the copy icon stays, and the button's label and tooltip
	// say the copy failed for a moment instead of faking a success.
	private _showFailure(): void {
		const btn = this.btnClicked;
		if (!btn || this.isAnimationRunning === true) return;

		this.isAnimationRunning = true;
		const label = btn.getAttribute("aria-label");
		const title = btn.getAttribute("title");
		btn.setAttribute("aria-label", CopyController.FAILED_TEXT);
		btn.setAttribute("title", CopyController.FAILED_TEXT);
		this._announce(CopyController.FAILED_TEXT);
		window.setTimeout(() => {
			if (label === null) btn.removeAttribute("aria-label");
			else btn.setAttribute("aria-label", label);
			if (title === null) btn.removeAttribute("title");
			else btn.setAttribute("title", title);
			this.isAnimationRunning = false;
		}, CopyController.FAILURE_DELAY);
	}

	// utility method to write to clipboard
	private async _writeToClipboard(
		text: string,
		icons: HTMLElement[],
	): Promise<void> {
		if (!navigator.clipboard) {
			console.error("Copy: Clipboard API is not supported in this browser.");
			this._showFailure();
			return;
		}

		try {
			await navigator.clipboard.writeText(text);
			this._showFeedback(icons);
		} catch (err) {
			console.error("Copy: Error while copying text.", err);
			this._showFailure();
		}
	}

	// copy text to clipboard
	private async _copyTextToClipboard(
		text: string,
		icons: HTMLElement[],
	): Promise<void> {
		let cleaned = text.trim();
		if (cleaned.startsWith("/")) {
			cleaned = window.location.origin + cleaned;
		}
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

	// Fetch + copy. Used by buttons that need the source bytes lazily
	// (e.g. the state explorer's "Copy package JSON" - server no longer
	// inlines the raw payload in the page, so this fetches ?state&json
	// on click). Same-origin enforced explicitly: defense-in-depth if a
	// future caller renders an attacker-controlled URL into the attribute.
	private async _fetchAndCopyToClipboard(
		url: string,
		icons: HTMLElement[],
	): Promise<void> {
		try {
			const target = new URL(url, location.href);
			if (target.origin !== location.origin) {
				console.error(`Copy: refusing cross-origin fetch ${target.origin}`);
				this._showFailure();
				return;
			}
			// TODO: show a loading state (idle → fetching → success/error) during
			// the fetch — to be addressed in a dedicated PR alongside a consistent
			// async-action feedback pattern for the explorer (copy, htmx triggers,
			// pagination etc.).
			const res = await fetch(target.toString(), {
				credentials: "same-origin",
			});
			if (!res.ok) {
				console.error(
					`Copy: fetch ${target.toString()} returned ${res.status}`,
				);
				this._showFailure();
				return;
			}
			const text = await res.text();
			await this._writeToClipboard(text, icons);
		} catch (err) {
			console.error("Copy: fetch failed.", err);
			this._showFailure();
		}
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

		// Handle data-copy-fetch (lazy same-origin fetch of bytes).
		if (this.getValue("fetch")) {
			this._fetchAndCopyToClipboard(this.getValue("fetch"), btnClickedIcons);
			return;
		}

		console.warn("Copy: No content to copy found on the button.");
	}
}
