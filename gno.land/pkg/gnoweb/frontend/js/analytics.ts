// SimpleAnalytics event delegation. The inline script in analytics.html sets
// window.sa_metadata before SA's latest.js fires; this file only attaches the
// click and submit listeners that forward custom events.

declare global {
	interface Window {
		sa_event?: (name: string, meta?: Record<string, string>) => void;
	}
}

const MAX_FUNC_NAME = 64;

function fire(name: string, meta: Record<string, string>): void {
	window.sa_event?.(name, meta);
}

// Copy button → copy_action. Capture phase so we fire before any controller
// that might stopPropagation on the click.
document.addEventListener(
	"click",
	(ev) => {
		if (!(ev.target instanceof Element)) return;
		const btn = ev.target.closest<HTMLButtonElement>(
			'button[data-controller~="copy"]',
		);
		if (!btn) return;
		let kind = "unknown";
		if (btn.hasAttribute("data-copy-text-value")) kind = "link";
		else if (btn.hasAttribute("data-copy-remote-value")) kind = "snippet";
		fire("copy_action", { kind });
	},
	true,
);

// Action form submission → submit_action. Capture phase is required:
// FormExecController._handleSubmit calls stopPropagation on bubble.
document.addEventListener(
	"submit",
	(ev) => {
		if (!(ev.target instanceof Element)) return;
		const article = ev.target.closest("[data-action-function-name-value]");
		if (!article) return;
		// Realm authors set this attribute freely; cap to avoid unbounded cardinality.
		const raw = article.getAttribute("data-action-function-name-value") ?? "";
		fire("submit_action", { func: raw.slice(0, MAX_FUNC_NAME) });
	},
	true,
);

export {};
