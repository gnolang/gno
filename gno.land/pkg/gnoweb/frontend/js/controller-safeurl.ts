import { BaseController } from "./controller.js";

/**
 * SafeurlController handles polling for pending SafeURL scan results
 * and updates the UI when scans complete.
 */
export class SafeurlController extends BaseController {
	static controllerIdentifier = "safeurl";
	static POLL_INTERVAL = 2000; // 2 seconds
	static MAX_POLLS = 30; // Max 1 minute of polling

	protected connect(): void {
		this.pollPendingScans();
	}

	/**
	 * Find all elements with pending SafeURL scans and poll for their status.
	 */
	private pollPendingScans(): void {
		const pendingElements = this.element.querySelectorAll(
			'[data-safeurl-status="pending"]',
		);
		if (pendingElements.length === 0) return;

		pendingElements.forEach((el) => {
			const scanId = el.getAttribute("data-safeurl-scan-id");
			if (scanId) {
				this.pollScanStatus(scanId, el as HTMLElement, 0);
			}
		});
	}

	/**
	 * Poll the API for scan status and update the element when complete.
	 */
	private async pollScanStatus(
		scanId: string,
		element: HTMLElement,
		pollCount: number,
	): Promise<void> {
		if (pollCount >= SafeurlController.MAX_POLLS) {
			this.markUnavailable(element);
			return;
		}

		try {
			const response = await fetch(`/api/safeurl/scan/${scanId}`);
			if (!response.ok) {
				this.markUnavailable(element);
				return;
			}

			const result = await response.json();

			if (result.status === "pending") {
				// Still pending, poll again after interval
				setTimeout(() => {
					this.pollScanStatus(scanId, element, pollCount + 1);
				}, SafeurlController.POLL_INTERVAL);
				return;
			}

			// Scan complete - update the element
			this.updateElement(element, result);
		} catch (error) {
			console.error("SafeURL poll error:", error);
			this.markUnavailable(element);
		}
	}

	/**
	 * Update an element with the final scan result.
	 */
	private updateElement(
		element: HTMLElement,
		result: { status: string; verdict?: string },
	): void {
		element.setAttribute("data-safeurl-status", result.status);
		if (result.verdict) {
			element.setAttribute("data-safeurl-verdict", result.verdict);
		}

		// Check element types BEFORE removing classes
		const isImagePlaceholder = element.classList.contains("img-placeholder");
		const isPendingLink = element.classList.contains("link-pending");

		// Remove scanning indicator
		const spinner = element.querySelector(
			".spinning, .link-scanning, .img-scanning",
		);
		if (spinner) {
			spinner.remove();
		}

		// Remove placeholder text if present
		const placeholderText = element.querySelector(".img-placeholder-text");
		if (placeholderText) {
			placeholderText.remove();
		}

		// Remove copy button for pending links
		const copyBtn = element.querySelector(".link-copy-btn");
		if (copyBtn) {
			copyBtn.remove();
		}

		// Update classes based on status
		element.classList.remove("link-pending", "img-pending", "img-placeholder");

		if (result.status === "safe") {
			// For image placeholders, load the actual image
			if (isImagePlaceholder) {
				this.loadSafeImage(element);
			} else if (isPendingLink) {
				// Convert pending link span to clickable anchor
				this.convertToLink(element, result.verdict);
			} else if (result.verdict) {
				const titleEl = element.querySelector("[title]") || element;
				const currentTitle = titleEl.getAttribute("title") || "";
				const newTitle = currentTitle
					? `${currentTitle} | SafeURL: ${result.verdict}`
					: `SafeURL: ${result.verdict}`;
				titleEl.setAttribute("title", newTitle);
			}
		} else if (result.status === "unsafe") {
			if (isImagePlaceholder) {
				this.showUnsafeImageWarning(element);
			} else if (isPendingLink) {
				// Convert to unsafe link with warning
				this.convertToUnsafeLink(element);
			} else {
				element.classList.add("link-unsafe", "img-unsafe");
				this.addWarningIcon(element);
			}
		} else {
			// unavailable or other
			this.markUnavailable(element);
		}
	}

	/**
	 * Convert a pending link span to an unsafe (but clickable) anchor element.
	 */
	private convertToUnsafeLink(element: HTMLElement): void {
		const href = element.getAttribute("data-safeurl-href");
		if (!href) return;

		const anchor = document.createElement("a");
		anchor.href = href;
		anchor.textContent = this.extractLinkText(element);
		anchor.setAttribute("rel", "noopener nofollow ugc");
		anchor.setAttribute("data-safeurl-status", "unsafe");

		// Add warning icon with tooltip attributes (matching server-rendered structure)
		const warningSpan = document.createElement("span");
		warningSpan.className = "link-unsafe-indicator tooltip";
		warningSpan.setAttribute("data-tooltip-target", "info");
		warningSpan.setAttribute("data-tooltip", "This link may be unsafe");
		warningSpan.setAttribute("title", "This link may be unsafe");
		warningSpan.innerHTML =
			'<svg class="c-icon"><use href="#ico-warning"></use></svg>';
		anchor.appendChild(warningSpan);

		// Replace the span with the anchor
		element.replaceWith(anchor);
	}

	/**
	 * Convert a pending link span to a clickable anchor element.
	 */
	private convertToLink(element: HTMLElement, verdict?: string): void {
		const href = element.getAttribute("data-safeurl-href");
		if (!href) return;

		const anchor = document.createElement("a");
		anchor.href = href;
		anchor.textContent = this.extractLinkText(element);
		anchor.setAttribute("rel", "noopener nofollow ugc");
		anchor.setAttribute("data-safeurl-status", "safe");
		if (verdict) {
			anchor.setAttribute("data-safeurl-verdict", verdict);
			anchor.title = `SafeURL: ${verdict}`;
		}

		// Add external link icon with tooltip attributes (matching server-rendered structure)
		const iconSpan = document.createElement("span");
		iconSpan.className = "link-external tooltip";
		iconSpan.setAttribute("data-tooltip-target", "info");
		iconSpan.setAttribute("data-tooltip", "External link");
		iconSpan.setAttribute("title", "External link");
		iconSpan.innerHTML =
			'<svg class="c-icon"><use href="#ico-external-link"></use></svg>';
		anchor.appendChild(iconSpan);

		// Replace the span with the anchor
		element.replaceWith(anchor);
	}

	/**
	 * Load the actual image for a safe image placeholder.
	 */
	private loadSafeImage(element: HTMLElement): void {
		const src = element.getAttribute("data-safeurl-src");
		const alt = element.getAttribute("data-safeurl-alt") || "";
		const verdict = element.getAttribute("data-safeurl-verdict") || "";

		if (!src) return;

		const img = document.createElement("img");
		img.src = src;
		img.alt = alt;
		if (verdict) {
			img.title = `SafeURL: ${verdict}`;
		}
		img.setAttribute("data-safeurl-status", "safe");
		if (verdict) {
			img.setAttribute("data-safeurl-verdict", verdict);
		}

		// Replace placeholder with actual image
		element.replaceWith(img);
	}

	/**
	 * Show warning for unsafe image placeholder.
	 */
	private showUnsafeImageWarning(element: HTMLElement): void {
		const alt = element.getAttribute("data-safeurl-alt") || "Image";
		element.innerHTML = `
      <svg class="c-icon"><use href="#ico-warning"></use></svg>
      <span class="img-unsafe-text">Unsafe image blocked: ${alt}</span>
    `;
		element.classList.add("img-unsafe");
		element.title = "This image was blocked because it may be unsafe";
	}

	/**
	 * Mark an element as unavailable (scan failed or timed out).
	 */
	private markUnavailable(element: HTMLElement): void {
		element.setAttribute("data-safeurl-status", "unavailable");
		element.classList.remove("link-pending", "img-pending");
		element.classList.add("link-unavailable", "img-unavailable");

		// Remove scanning indicator
		const spinner = element.querySelector(
			".spinning, .link-scanning, .img-scanning",
		);
		if (spinner) {
			spinner.remove();
		}
	}

	/**
	 * Add a warning icon to unsafe elements.
	 */
	private addWarningIcon(element: HTMLElement): void {
		const warning = document.createElement("span");
		warning.className = "link-unsafe-indicator";
		warning.setAttribute("title", "This link may be unsafe");
		warning.innerHTML =
			'<svg class="c-icon"><use href="#ico-warning"></use></svg>';
		element.appendChild(warning);
	}

	/**
	 * Extract text content from an element, excluding buttons and other non-text children.
	 */
	private extractLinkText(element: HTMLElement): string {
		let text = "";
		for (const node of element.childNodes) {
			if (node.nodeType === Node.TEXT_NODE) {
				text += node.textContent;
			}
		}
		return text.trim();
	}
}
