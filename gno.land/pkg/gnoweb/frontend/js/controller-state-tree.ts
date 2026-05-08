import { BaseController } from "./controller.js";

// StateTreeController persists which `<details data-tree-key="...">`
// rows the user has opened, scoped per realm via
// `data-state-tree-pkg-value`.
//
// Storage is localStorage rather than a cookie because:
//   - Volume: a chatty realm can have hundreds of OIDs; a cookie of
//     that size ships on every request.
//   - Server has no opinion to express here — the default render is
//     always "top-level open, the rest closed", and JS reconciles
//     on connect.
//
// Storage shape: `state_tree_open:<pkgPath>` → `["OID1","OID2",…]`.
// Per-realm scoping prevents Boards' open-state from polluting
// Gnoswap's, etc.
//
// Sidebar TOC scroll is purely CSS+anchor: each top-level row has an
// `id` with a per-view suffix (`-pretty` / `-tree`), the sidebar
// emits two anchors, and CSS hides the inactive one. No JS bridge.
export class StateTreeController extends BaseController {
	private storageKey = "";
	private openSet = new Set<string>();

	protected connect(): void {
		const pkg = this.getValue("pkg") || "global";
		this.storageKey = `state_tree_open:${pkg}`;
		this.openSet = this.loadOpen();
		this.applyOpen();

		// Listen for `toggle` events bubbling from any <details> in
		// the tree. One listener for the whole subtree — cheaper than
		// one-per-details and handles dynamically-added nodes too.
		// `toggle` doesn't bubble in older specs, so we capture.
		this.element.addEventListener(
			"toggle",
			this.onToggle as EventListener,
			true,
		);
	}

	private onToggle = (event: Event): void => {
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
		for (const key of this.openSet) {
			const sel = `details[data-tree-key="${cssEscape(key)}"]`;
			const el = this.element.querySelector<HTMLDetailsElement>(sel);
			if (el && !el.open) el.open = true;
		}
	}

	private loadOpen(): Set<string> {
		try {
			const raw = localStorage.getItem(this.storageKey);
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
				this.storageKey,
				JSON.stringify(Array.from(this.openSet)),
			);
		} catch {
			// localStorage unavailable / quota — silent skip.
		}
	}
}

// cssEscape: minimal shim around CSS.escape (which is widely supported
// but absent in some test environments). Falls back to a quote-safe
// substring so attribute selectors don't break on `:` (common in OIDs).
function cssEscape(s: string): string {
	if (typeof CSS !== "undefined" && typeof CSS.escape === "function") {
		return CSS.escape(s);
	}
	return s.replace(/"/g, '\\"');
}
