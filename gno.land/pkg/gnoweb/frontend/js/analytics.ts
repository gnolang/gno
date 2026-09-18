// SimpleAnalytics event delegation. sa-bootstrap.js sets window.sa_metadata
// synchronously before SA's latest.js loads; this file attaches all
// click/submit/change/scroll listeners that forward custom events.
// Window.sa_event / sa_metadata types are declared in events.d.ts.

const MAX_FUNC_NAME = 64;
const MAX_PKGPATH = 128;
const SCROLL_THRESHOLDS = [50, 75, 100] as const;

function fire(name: string, meta?: Record<string, string | boolean>): void {
	window.sa_event?.(name, meta);
}

// Delegate a document-level click to the nearest ancestor matching `selector`.
// `capture` is required for handlers that must run before a controller calls
// stopPropagation on the bubble phase (e.g. copy/submit).
function onClick<T extends Element>(
	selector: string,
	handler: (el: T) => void,
	capture = false,
): void {
	document.addEventListener(
		"click",
		(ev) => {
			if (!(ev.target instanceof Element)) return;
			const match = ev.target.closest<T>(selector);
			if (match) handler(match);
		},
		capture,
	);
}

// ---- copy_action: split kind by data-copy-* attribute pattern.
// Capture phase so we fire before any controller stops propagation.
onClick<HTMLButtonElement>(
	'button[data-controller~="copy"]',
	(btn) => {
		const remote = btn.getAttribute("data-copy-remote-value") ?? "";
		let kind = "unknown";
		if (btn.hasAttribute("data-copy-text-value")) kind = "link";
		else if (remote === "source-code") kind = "source";
		else if (remote.startsWith("func-")) kind = "func_signature";
		else if (remote.startsWith("action-function-")) kind = "gnokey_command";
		fire("copy_action", { kind });
	},
	true,
);

// ---- submit_action: action-form submission with func + pkgpath.
// Capture phase: FormExecController.stopPropagation runs on bubble.
document.addEventListener(
	"submit",
	(ev) => {
		if (!(ev.target instanceof Element)) return;
		const article = ev.target.closest("[data-action-function-name-value]");
		if (!article) return;
		const name = article.getAttribute("data-action-function-name-value") ?? "";
		const pkg =
			article.getAttribute("data-action-function-pkgpath-value") ?? "";
		fire("submit_action", {
			func: name.slice(0, MAX_FUNC_NAME),
			pkgpath: pkg.slice(0, MAX_PKGPATH),
		});
	},
	true,
);

// ---- search_action: header searchbar form submit. Count only, no query text.
document.addEventListener("submit", (ev) => {
	if (ev.target instanceof Element && ev.target.matches("form.searchbar")) {
		fire("search_action");
	}
});

// ---- breadcrumb_click: anchor clicks inside the breadcrumb list.
onClick('ol[data-searchbar-target="breadcrumb"] a', () => {
	fire("breadcrumb_click");
});

// ---- back_navigation: browser back/forward.
window.addEventListener("popstate", () => fire("back_navigation"));

// ---- mode_change: action-header dispatches "mode:changed" with the chosen
// mode (typed via DocumentEventMap augmentation in events.d.ts).
document.addEventListener("mode:changed", (ev) => {
	const mode = ev.detail?.mode;
	if (typeof mode === "string") fire("mode_change", { mode });
});

// ---- send_mode_toggle: click on the labels that wrap the send-mode checkbox.
onClick<HTMLElement>(
	'label[data-action="click->action-function#updateAllFunctionsSend"]',
	(label) => {
		const param = label.getAttribute("data-action-function-send-param");
		fire("send_mode_toggle", { active: param === "true" });
	},
);

// ---- theme_toggle: theme controller dispatches "theme:changed" with the
// user-chosen preference (light, dark, or system) — never the resolved theme,
// so picking "system" is distinguishable from explicit light/dark.
document.addEventListener("theme:changed", (ev) => {
	const theme = ev.detail?.theme;
	if (typeof theme === "string") fire("theme_toggle", { theme });
});

// ---- network_popup_toggle: change on the popup checkbox.
const networkToggle = document.querySelector<HTMLInputElement>(
	"#searchbar-server-popup-toggle",
);
if (networkToggle) {
	networkToggle.addEventListener("change", () => {
		fire("network_popup_toggle", { open: networkToggle.checked });
	});
}

// ---- devmode_toggle: change on the dev-menu checkbox (homepage only).
const devmodeToggle = document.querySelector<HTMLInputElement>(
	"#header-input-devmode",
);
if (devmodeToggle) {
	devmodeToggle.addEventListener("change", () => {
		fire("devmode_toggle", { enabled: devmodeToggle.checked });
	});
}

// ---- toc_toggle: native <details> toggle event does NOT bubble; attach
// directly to each accordion present at load time.
document
	.querySelectorAll<HTMLDetailsElement>("details.accordion")
	.forEach((d) => {
		d.addEventListener("toggle", () => {
			fire("toc_toggle", { open: d.open });
		});
	});

// ---- qeval_preview: watch the result element for an outcome change.
// Placeholder text and error class are read from data-* attributes set on the
// element by the server (action.html); the controller reads the same source,
// so analytics never holds an independent copy of these strings.
const qevalResult = document.querySelector<HTMLElement>(
	'[data-action-function-target="qeval-result"]',
);
if (qevalResult) {
	const placeholder = qevalResult.dataset.qevalPlaceholder ?? "";
	const errorClass = qevalResult.dataset.qevalErrorClass ?? "u-color-danger";
	let lastSuccess: boolean | null = null;
	const observer = new MutationObserver(() => {
		const text = qevalResult.textContent?.trim() ?? "";
		if (text === "" || text === placeholder) return;
		const success = !qevalResult.classList.contains(errorClass);
		if (success !== lastSuccess) {
			lastSuccess = success;
			fire("qeval_preview", { success });
		}
	});
	observer.observe(qevalResult, {
		attributes: true,
		attributeFilter: ["class"],
		childList: true,
		subtree: true,
		characterData: true,
	});
}

// ---- address_filled / params_filled: fire once per page-load when any value
// becomes non-empty.
const addressInput = document.querySelector<HTMLInputElement>(
	"#action-user-address",
);
if (addressInput) {
	const onInput = (): void => {
		if (!addressInput.value.trim()) return;
		addressInput.removeEventListener("input", onInput);
		fire("address_filled");
	};
	addressInput.addEventListener("input", onInput);
}

const paramInputs = document.querySelectorAll<HTMLInputElement>(
	'[data-action-function-target="param-input"]',
);
if (paramInputs.length > 0) {
	const onInput = (ev: Event): void => {
		if (!(ev.target instanceof HTMLInputElement)) return;
		if (!ev.target.value.trim()) return;
		paramInputs.forEach((i) => i.removeEventListener("input", onInput));
		fire("params_filled");
	};
	paramInputs.forEach((input) => input.addEventListener("input", onInput));
}

// ---- list_filter_search: debounced input on the user-page filter.
// Count only — never the query text.
const filterInput =
	document.querySelector<HTMLInputElement>("#packages-search");
if (filterInput) {
	let timer: ReturnType<typeof setTimeout> | undefined;
	filterInput.addEventListener("input", () => {
		if (timer !== undefined) clearTimeout(timer);
		timer = setTimeout(() => fire("list_filter_search"), 250);
	});
}

// ---- list_sort_change: radio-button change on order-mode.
document
	.querySelectorAll<HTMLInputElement>('input[name="order-mode"]')
	.forEach((input) => {
		input.addEventListener("change", () => {
			if (input.checked) fire("list_sort_change", { order: input.value });
		});
	});

// ---- list_display_change: radio-button change on display-mode.
document
	.querySelectorAll<HTMLInputElement>('input[name="display-mode"]')
	.forEach((input) => {
		input.addEventListener("change", () => {
			if (input.checked) fire("list_display_change", { mode: input.value });
		});
	});

// ---- scroll_depth: window scroll on source/action pages, fires once per
// threshold per page-load. The server-side page_type "help" identifies the
// action page; remap it to "action" for a clearer surface label.
const surface =
	window.sa_metadata?.page_type === "source"
		? "source"
		: window.sa_metadata?.page_type === "help"
			? "action"
			: null;
if (surface) {
	const fired = new Set<number>();
	const onScroll = () => {
		const scrolled = window.scrollY + window.innerHeight;
		const total = document.documentElement.scrollHeight;
		if (total <= 0) return;
		const pct = (scrolled / total) * 100;
		for (const t of SCROLL_THRESHOLDS) {
			if (pct >= t && !fired.has(t)) {
				fired.add(t);
				fire("scroll_depth", { threshold: String(t), surface });
			}
		}
		if (fired.size === SCROLL_THRESHOLDS.length) {
			window.removeEventListener("scroll", onScroll);
		}
	};
	window.addEventListener("scroll", onScroll, { passive: true });
}

// ---- outbound_<target>: anchors carrying data-outbound fire a named event in
// addition to SA's generic outbound auto-event. Target value is server-rendered
// from a fixed enum (see SIMPLEANALYTICS.md) so cardinality stays bounded.
onClick<HTMLAnchorElement>("a[data-outbound]", (a) => {
	const target = a.getAttribute("data-outbound");
	if (target) fire(`outbound_${target}`);
});

export {};
