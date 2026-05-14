import "htmx.org";
import { BaseController } from "../../../frontend/js/controller.js";

// Client-only view-mode pref (no SSR cookie — cache key stays URL-only).
const VIEW_STORAGE_KEY = "stateViewMode";
const VIEW_VALID_MODES = new Set(["pretty", "tree"]);

export class StateController extends BaseController {
	private declare treeStorageKey: string;
	private declare openSet: Set<string>;
	private declare viewTree: HTMLElement | null;
	private docIndex: Record<string, string> = {};
	private bulkInProgress = false;

	protected connect(): void {
		this.viewTree = this.getTarget("view-tree");
		this.treeStorageKey = `state_tree_open:${this.getValue("pkg") || "global"}`;
		this.openSet = this.loadOpen();
		this.loadDocIndex();
		if (this.viewTree) {
			this.applyOpen(this.viewTree);
			this.syncToggleAllState();
			this.projectDocs(this.viewTree);
			// Capture: older engines didn't bubble `toggle`.
			this.viewTree.addEventListener("toggle", this.onToggle, true);
		}
		this.restoreViewMode();
		document.addEventListener("htmx:afterSwap", this.onAfterSwap);
	}

	protected disconnect(): void {
		this.viewTree?.removeEventListener("toggle", this.onToggle, true);
		document.removeEventListener("htmx:afterSwap", this.onAfterSwap);
	}

	// Bounded to top-level (ADR-004 §5): recursive expand-all would
	// burst N fragment fetches and reintroduce amplification. The
	// top-level <details> are direct children of `.tree`, not of the
	// data-state-target host (`.view-tree`) — scope to `.tree` so the
	// `:scope > details` filter actually matches.
	public toggleAll(event: Event): void {
		if (!this.viewTree) return;
		const btn = event.currentTarget;
		if (!(btn instanceof HTMLElement)) return;
		const expand = btn.getAttribute("aria-pressed") !== "true";
		const tree =
			this.viewTree.querySelector<HTMLElement>(":scope > .tree") ??
			this.viewTree;
		const top = tree.querySelectorAll<HTMLDetailsElement>(":scope > details");
		this.bulkInProgress = true;
		for (const d of top) if (d.open !== expand) d.open = expand;
		this.bulkInProgress = false;
		if (expand) {
			for (const d of top) {
				const k = d.getAttribute("data-tree-key");
				if (k) this.openSet.add(k);
			}
		} else this.openSet.clear();
		this.persistOpen();
		btn.setAttribute("aria-pressed", String(expand));
		const label = this.getTarget("toggle-label");
		if (label) label.textContent = expand ? "Collapse all" : "Expand all";
	}

	// applyOpen() restores an arbitrary open-subset, so the template's
	// static aria-pressed can lie — re-derive from the live DOM.
	private syncToggleAllState(): void {
		if (!this.viewTree) return;
		const tree =
			this.viewTree.querySelector<HTMLElement>(":scope > .tree") ??
			this.viewTree;
		const top = tree.querySelectorAll<HTMLDetailsElement>(":scope > details");
		const label = this.getTarget("toggle-label");
		const btn = label?.closest("button");
		if (!btn || top.length === 0) return;
		const allOpen = [...top].every((d) => d.open);
		btn.setAttribute("aria-pressed", String(allOpen));
		if (label) label.textContent = allOpen ? "Collapse all" : "Expand all";
	}

	private onToggle = (e: Event): void => {
		if (this.bulkInProgress) return;
		const t = e.target;
		if (!(t instanceof HTMLDetailsElement)) return;
		const k = t.getAttribute("data-tree-key");
		if (!k) return;
		if (t.open) this.openSet.add(k);
		else this.openSet.delete(k);
		this.persistOpen();
	};

	// Scoped: re-applied only inside the htmx swap target.
	private applyOpen(root: HTMLElement): void {
		if (this.openSet.size === 0) return;
		for (const el of root.querySelectorAll<HTMLDetailsElement>(
			"details[data-tree-key]",
		)) {
			const k = el.getAttribute("data-tree-key");
			if (k && this.openSet.has(k) && !el.open) el.open = true;
		}
	}

	private onAfterSwap = (e: Event): void => {
		const root = (e as CustomEvent<{ target?: HTMLElement }>).detail?.target;
		if (!root) return;
		this.applyOpen(root);
		this.projectDocs(root);
		// Deep-link after expansion: scroll iff #anchor is now in DOM.
		const hash = window.location.hash.slice(1);
		if (!hash) return;
		root
			.querySelector(`#${CSS.escape(hash)}`)
			?.scrollIntoView({ behavior: "smooth", block: "start" });
	};

	// ADR-004 §8: doc-index island. Server inlines a `{name: doc}` JSON
	// blob keyed by top-level decl name. Parsed once at connect; the
	// projection runs again per htmx swap to hydrate lazy-loaded fragments.
	private loadDocIndex(): void {
		try {
			const el = document.getElementById("state-doc-index");
			if (!el?.textContent) return;
			const parsed = JSON.parse(el.textContent);
			if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
				this.docIndex = parsed as Record<string, string>;
			}
		} catch {
			// Template guarantees `{}` fallback — reaching here means the realm
			// shipped malformed doc data. Skip silently.
		}
	}

	// Populate empty `[data-doc-slot]` placeholders inside the swap target
	// when the enclosing `[data-name]` matches a doc-index entry. Idempotent:
	// slots that already carry text are left untouched (server-rendered
	// top-level decls win over the client projection).
	private projectDocs(root: HTMLElement): void {
		for (const named of root.querySelectorAll<HTMLElement>("[data-name]")) {
			const name = named.getAttribute("data-name");
			if (!name) continue;
			const doc = this.docIndex[name];
			if (!doc) continue;
			const slot = named.querySelector<HTMLElement>(":scope > [data-doc-slot]");
			if (!slot || slot.textContent?.trim()) continue;
			slot.textContent = doc;
		}
	}

	private loadOpen(): Set<string> {
		try {
			const raw = localStorage.getItem(this.treeStorageKey);
			const arr = raw ? JSON.parse(raw) : null;
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
				JSON.stringify([...this.openSet]),
			);
		} catch {}
	}

	// URL-driven view-mode (ADR-004 §6). CSS toggles via
	// `body:has(#state-view-tree:checked)` — the radio drives the swap.
	public updateView(event: Event): void {
		const t = event.target;
		if (!(t instanceof HTMLInputElement) || !t.checked) return;
		const mode = t.value;
		if (!VIEW_VALID_MODES.has(mode)) return;
		try {
			localStorage.setItem(VIEW_STORAGE_KEY, mode);
		} catch {}
		this.syncTocAria(mode);
		this.syncViewURL(mode);
	}

	// Must write to `$webargs`, not `?query` — the server's WebQuery only
	// reads from `$`, so `?view=tree` would silently fail to round-trip.
	private syncViewURL(mode: string): void {
		try {
			const next = setWebarg(
				window.location.href,
				"view",
				mode === "pretty" ? null : mode,
			);
			history.replaceState(null, "", next);
		} catch {
			// history API throws in sandboxed contexts; localStorage already saved.
		}
	}

	private syncTocAria(mode: string): void {
		const on = mode === "tree" ? "toc-tree" : "toc-pretty";
		const off = mode === "tree" ? "toc-pretty" : "toc-tree";
		for (const a of this.getTargets(on)) {
			a.removeAttribute("aria-hidden");
			a.removeAttribute("tabindex");
		}
		for (const a of this.getTargets(off)) {
			a.setAttribute("aria-hidden", "true");
			a.setAttribute("tabindex", "-1");
		}
	}

	// Flip the view-radio to `mode` (if not already there) and propagate
	// via the existing TOC + URL sync.
	private switchViewMode(mode: string): void {
		for (const r of this.getTargets("view-radio")) {
			if (r instanceof HTMLInputElement && r.value === mode && !r.checked) {
				r.checked = true;
				this.syncTocAria(mode);
				this.syncViewURL(mode);
				return;
			}
		}
	}

	// URL wins (server already rendered the matching mode); otherwise
	// honor localStorage and stamp the URL so it stays shareable.
	private restoreViewMode(): void {
		// Explicit deep-link wins over both URL webarg and localStorage:
		// `#<anchor>-pretty|-tree` lands on a CSS-hidden view otherwise.
		if (this.reconcileAnchorView()) return;
		let urlMode: string | null = null;
		try {
			urlMode = getWebarg(window.location.href, "view");
		} catch {}
		if (urlMode && VIEW_VALID_MODES.has(urlMode)) {
			try {
				localStorage.setItem(VIEW_STORAGE_KEY, urlMode);
			} catch {}
			return; // server-render already correct; nothing to do
		}
		let saved: string | null = null;
		try {
			saved = localStorage.getItem(VIEW_STORAGE_KEY);
		} catch {
			return;
		}
		if (!saved || !VIEW_VALID_MODES.has(saved) || saved === "pretty") return;
		// Local pref differs from server default — flip radio + stamp URL.
		this.switchViewMode(saved);
	}

	// A shared `#<anchor>-pretty|-tree` link can target the view CSS is
	// hiding. If so, switch the radio to the hash's view so the element
	// is visible and `:target` + scroll take over. Returns true if it
	// recognised a view-suffixed hash (caller then skips saved-pref).
	private reconcileAnchorView(): boolean {
		const hash = window.location.hash.slice(1);
		const dash = hash.lastIndexOf("-");
		if (dash === -1) return false;
		const mode = hash.slice(dash + 1);
		if (!VIEW_VALID_MODES.has(mode)) return false;
		this.switchViewMode(mode);
		try {
			localStorage.setItem(VIEW_STORAGE_KEY, mode);
		} catch {}
		return true;
	}
}

// gnoweb URLs are `<path>[$<webargs>][?<query>]`. URLSearchParams only
// sees `?query`, so we hand-split on `$`. Pairs keep the "bare key" form
// (e.g. `state` with no `=`) to round-trip the server's canonical shape.

function splitGnoURL(href: string): {
	prefix: string;
	webargs: string;
	suffix: string;
} {
	const u = new URL(href, window.location.href);
	const suffix = u.search + u.hash;
	const dollar = u.pathname.indexOf("$");
	if (dollar === -1) {
		return { prefix: u.origin + u.pathname, webargs: "", suffix };
	}
	return {
		prefix: u.origin + u.pathname.slice(0, dollar),
		webargs: u.pathname.slice(dollar + 1),
		suffix,
	};
}

function parseWebargs(webargs: string): Array<[string, string | null]> {
	if (!webargs) return [];
	const out: Array<[string, string | null]> = [];
	for (const part of webargs.split("&")) {
		if (!part) continue;
		const eq = part.indexOf("=");
		// Treat `key` and `key=` alike: the server's EncodeValues drops
		// `=` for empty values, so the bare form is canonical.
		if (eq === -1 || eq === part.length - 1) {
			out.push([
				decodeURIComponent(part.slice(0, eq === -1 ? undefined : eq)),
				null,
			]);
		} else {
			out.push([
				decodeURIComponent(part.slice(0, eq)),
				decodeURIComponent(part.slice(eq + 1)),
			]);
		}
	}
	return out;
}

function serializeWebargs(pairs: Array<[string, string | null]>): string {
	const parts: string[] = [];
	for (const [k, v] of pairs) {
		const key = encodeURIComponent(k);
		parts.push(v === null ? key : `${key}=${encodeURIComponent(v)}`);
	}
	return parts.join("&");
}

function getWebarg(href: string, key: string): string | null {
	for (const [k, v] of parseWebargs(splitGnoURL(href).webargs)) {
		if (k === key) return v ?? "";
	}
	return null;
}

function setWebarg(href: string, key: string, value: string | null): string {
	const { prefix, webargs, suffix } = splitGnoURL(href);
	const pairs = parseWebargs(webargs);
	const idx = pairs.findIndex(([k]) => k === key);
	if (value === null) {
		if (idx !== -1) pairs.splice(idx, 1);
	} else if (idx === -1) {
		pairs.push([key, value]);
	} else {
		pairs[idx] = [key, value];
	}
	const serialized = serializeWebargs(pairs);
	return prefix + (serialized ? "$" + serialized : "") + suffix;
}
