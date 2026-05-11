import { BaseController, setPrefCookie } from "./controller.js";

// localStorage = authoritative client pref; cookie = SSR helper so
// the server stamps the right radio `checked` from first paint.
const VIEW_STORAGE_KEY = "stateViewMode";
const VIEW_COOKIE_KEY = "state_view_mode";
const COOKIE_MAX_AGE = 365 * 24 * 60 * 60;
const VIEW_VALID_MODES = new Set(["pretty", "tree"]);

// State explorer feature controller. Three state-specific behaviors:
// (1) tree open-state localStorage per realm, (2) toggle-all, (3)
// Pretty/Tree view mode persistence. Free-text search lives in the
// agnostic `controller-search` primitive.
//
// Mounted on a `display: contents` wrapper that spans sidebar +
// header + subheader + article so setupActions reaches every
// `data-action="...->state#..."` across the page.
export class StateController extends BaseController {
	private declare treeStorageKey: string;
	private declare openSet: Set<string>;
	private declare viewTree: HTMLElement | null;
	private bulkInProgress = false;

	protected connect(): void {
		this.viewTree = this.element.querySelector<HTMLElement>(".view-tree");

		const pkg = this.getValue("pkg") || "global";
		this.treeStorageKey = `state_tree_open:${pkg}`;
		this.openSet = this.loadOpen();
		if (this.viewTree) {
			this.applyOpen();
			// Capture phase: cheap defense for older engines where `toggle`
			// historically didn't bubble (modern browsers do bubble it).
			this.viewTree.addEventListener(
				"toggle",
				this.onToggle as EventListener,
				true,
			);
		}

		this.restoreViewMode();
	}

	protected disconnect(): void {
		if (this.viewTree) {
			this.viewTree.removeEventListener(
				"toggle",
				this.onToggle as EventListener,
				true,
			);
		}
	}

	// Bulk expand/collapse. The cascading `toggle` events flow through
	// `onToggle` below so localStorage stays in sync — one source of
	// persistence truth, two surfaces (manual click + bulk).
	public toggleAll(event: Event): void {
		if (!this.viewTree) return;
		const btn = event.currentTarget;
		if (!(btn instanceof HTMLElement)) return;
		const expand = btn.getAttribute("aria-pressed") !== "true";
		const details =
			this.viewTree.querySelectorAll<HTMLDetailsElement>("details");
		// Gate onToggle reconciliation so we write localStorage once
		// instead of N times during the burst.
		this.bulkInProgress = true;
		for (const d of details) {
			if (d.open !== expand) d.open = expand;
		}
		this.bulkInProgress = false;
		if (expand) {
			for (const d of details) {
				const key = d.getAttribute("data-tree-key");
				if (key) this.openSet.add(key);
			}
		} else {
			this.openSet.clear();
		}
		this.persistOpen();
		btn.setAttribute("aria-pressed", String(expand));
		const label = this.getTarget("toggle-label");
		if (label) label.textContent = expand ? "Collapse all" : "Expand all";
	}

	private onToggle = (event: Event): void => {
		if (this.bulkInProgress) return;
		const target = event.target as HTMLElement;
		if (!(target instanceof HTMLDetailsElement)) return;
		const key = target.getAttribute("data-tree-key");
		if (!key) return;
		if (target.open) {
			this.openSet.add(key);
		} else {
			this.openSet.delete(key);
		}
		this.persistOpen();
	};

	private applyOpen(): void {
		if (!this.viewTree || this.openSet.size === 0) return;
		// Single querySelectorAll + Set lookup — O(N), avoids N
		// per-key queries on chatty realms.
		const all = this.viewTree.querySelectorAll<HTMLDetailsElement>(
			"details[data-tree-key]",
		);
		for (const el of all) {
			const key = el.getAttribute("data-tree-key");
			if (key && this.openSet.has(key) && !el.open) el.open = true;
		}
	}

	private loadOpen(): Set<string> {
		try {
			const raw = localStorage.getItem(this.treeStorageKey);
			if (!raw) return new Set();
			const arr = JSON.parse(raw);
			if (!Array.isArray(arr)) return new Set();
			return new Set(arr.filter((v): v is string => typeof v === "string"));
		} catch {
			return new Set();
		}
	}

	private persistOpen(): void {
		try {
			localStorage.setItem(
				this.treeStorageKey,
				JSON.stringify(Array.from(this.openSet)),
			);
		} catch {}
	}

	// Writes both stores: localStorage (client truth) + cookie (SSR
	// hint so the server stamps the right radio on next nav).
	public updateView(event: Event): void {
		const target = event.target;
		if (!(target instanceof HTMLInputElement) || !target.checked) return;
		const mode = target.value;
		if (!VIEW_VALID_MODES.has(mode)) return;
		try {
			localStorage.setItem(VIEW_STORAGE_KEY, mode);
		} catch {}
		setPrefCookie(VIEW_COOKIE_KEY, mode, COOKIE_MAX_AGE);
	}

	// Safety net: if the SSR cookie was stripped (privacy ext, proxy)
	// but localStorage survived, reconcile the radio on connect.
	private restoreViewMode(): void {
		let saved: string | null = null;
		try {
			saved = localStorage.getItem(VIEW_STORAGE_KEY);
		} catch {
			return;
		}
		if (!saved || !VIEW_VALID_MODES.has(saved)) return;
		const radios = this.getTargets("view-radio");
		for (const radio of radios) {
			if (
				radio instanceof HTMLInputElement &&
				radio.value === saved &&
				!radio.checked
			) {
				radio.checked = true;
				return;
			}
		}
	}
}
