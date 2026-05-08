import { BaseController } from "./controller.js";

// Storage key for the user's last-picked Pretty/Raw choice. Mirrors
// `controller-theme`'s split: localStorage holds the preference (the
// authoritative source the user controls), the cookie is a derived
// SSR helper so the server renders the right radio `checked` from
// first paint without a JS-driven flicker.
const STORAGE_KEY = "stateViewMode";
const COOKIE_KEY = "state_view_mode";
const COOKIE_MAX_AGE = 365 * 24 * 60 * 60; // 1 year in seconds
const VALID_MODES = new Set(["pretty", "tree"]);

// StateViewController persists the Pretty/Raw choice across reloads
// and history navigation. Mirrors `controller-theme`'s pattern:
// - localStorage is the authoritative preference (read on connect,
//   written on every change).
// - A companion cookie carries the same value to the server so SSR
//   stamps `checked` on the matching radio from first paint —
//   eliminates the Pretty→Raw flicker the async controller would
//   otherwise cause on a Raw-saved reload.
//
// The CSS toggle (radios driving `body:has(:checked)`) handles live
// in-page switching; the controller is purely for persistence.
export class StateViewController extends BaseController {
	protected connect(): void {
		this._restoreMode();
	}

	// _restoreMode reconciles the radio state with the saved choice
	// from localStorage. In the happy path the server already rendered
	// the right radio `checked` (via the cookie), so this is a no-op.
	// It still runs as a safety net: if the cookie was stripped (privacy
	// extension, server proxy) but localStorage is intact, the page
	// still ends up in the saved view.
	private _restoreMode(): void {
		let saved: string | null = null;
		try {
			saved = localStorage.getItem(STORAGE_KEY);
		} catch {
			return; // localStorage unavailable — preference can't be read
		}
		if (!saved || !VALID_MODES.has(saved)) return;
		const radios = this.getTargets("radio") as HTMLInputElement[];
		for (const radio of radios) {
			if (radio.value === saved && !radio.checked) {
				radio.checked = true;
				return;
			}
		}
	}

	// updateMode writes both stores on every user-driven change.
	// localStorage is the authoritative pref; the cookie keeps the
	// server in sync for the next page render.
	public updateMode(event: Event): void {
		const target = event.target as HTMLInputElement;
		if (!target.checked) return;
		const mode = target.value;
		if (!VALID_MODES.has(mode)) return;
		try {
			localStorage.setItem(STORAGE_KEY, mode);
		} catch {
			// localStorage unavailable — preference won't persist client-side,
			// but the cookie still lets the server render correctly on next nav.
		}
		const secure = location.protocol === "https:" ? ";Secure" : "";
		document.cookie = `${COOKIE_KEY}=${mode};path=/;max-age=${COOKIE_MAX_AGE};SameSite=Lax${secure}`;
	}
}
