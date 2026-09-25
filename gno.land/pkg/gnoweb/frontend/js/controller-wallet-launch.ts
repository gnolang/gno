import { BaseController } from "./controller.js";

// Entry shape of the server-embedded registry (components/wallets.json).
// platforms/install_url are informational for now (no filtering/fallback yet).
interface Wallet {
	name: string;
	id: string;
	icon: string;
	scheme: string; // bare, e.g. "land.gno.gnokey"; this controller appends "://tx?..."
	platforms: string[];
	install_url: string;
}

// The registry is parsed lazily on first submit and shared across the
// per-function controller instances.
let registryCache: Wallet[] | undefined;

// WalletLaunchController routes the Execute submit to an external wallet via a
// GnoConnect launch link, on mobile only. In every other case (desktop, no
// registered wallet, in-page extension present) the native submit proceeds.
export class WalletLaunchController extends BaseController {
	declare _funcName: string;
	declare _pkgPath: string;

	protected connect(): void {
		this.initializeDOM({});

		// Attached to the params <form> (one controller per element); the
		// function name/pkgpath live on the enclosing article.
		const article = this.element.closest<HTMLElement>(
			"[data-action-function-name-value]",
		);
		this._funcName =
			article?.getAttribute("data-action-function-name-value") || "";
		this._pkgPath =
			article?.getAttribute("data-action-function-pkgpath-value") || "";

		this.element.addEventListener("submit", this._onSubmit.bind(this));
	}

	// A missing/malformed registry disables external-wallet routing.
	private _wallets(): Wallet[] {
		if (registryCache) return registryCache;
		registryCache = [];
		const script = this.getGlobalTarget("wallet-registry");
		if (!script?.textContent) return registryCache;
		try {
			const parsed = JSON.parse(script.textContent);
			if (Array.isArray(parsed)) registryCache = parsed as Wallet[];
		} catch {
			this.warn("invalid wallet registry JSON");
		}
		return registryCache;
	}

	// Parameter name/value pairs read live from the inputs at submit time
	// (checked boxes of a checkbox group are comma-joined).
	private _readArgs(): Map<string, string> {
		const values = new Map<string, string>();
		this.element
			.querySelectorAll<HTMLInputElement>("[data-action-function-param-value]")
			.forEach((input) => {
				const name =
					input.getAttribute("data-action-function-param-value") || "";
				if (!name) return;
				if (input.type === "checkbox" || input.type === "radio") {
					const prev = values.get(name) ?? "";
					if (input.checked) {
						values.set(
							name,
							prev ? `${prev},${input.value.trim()}` : input.value.trim(),
						);
					} else if (!values.has(name)) {
						values.set(name, "");
					}
				} else {
					values.set(name, input.value.trim());
				}
			});
		return values;
	}

	// Send coins, if the send checkbox is toggled on.
	private _readSend(): string | undefined {
		const box = this.element.querySelector<HTMLInputElement>(
			'input[type="checkbox"][data-action-function-send-value]',
		);
		if (box?.checked) {
			return box.getAttribute("data-action-function-send-value") || undefined;
		}
		return undefined;
	}

	private _meta(name: string): string {
		const el = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`);
		return el?.content?.trim() || "";
	}

	// Coarse primary pointer only: maxTouchPoints would also match touchscreen
	// laptops, where a failed custom-scheme launch would break Execute.
	private _isMobile(): boolean {
		return window.matchMedia?.("(pointer: coarse)").matches === true;
	}

	// An in-page provider (browser extension) owns the submit; never intercept.
	private _hasInPageProvider(): boolean {
		const w = window as unknown as Record<string, unknown>;
		return Boolean(w.adena || w.gnoconnect);
	}

	// Current page URL minus wallet result params, so repeated round trips
	// don't accumulate stale status/hash.
	private _callbackURL(): string {
		const url = new URL(window.location.href);
		url.searchParams.delete("status");
		url.searchParams.delete("hash");
		return url.toString();
	}

	// Compose "<scheme>://tx?path=&func=&arg.<name>=&...". Args are named,
	// prefixed "arg." so realm parameter names can't collide with the link's
	// own keys (path, func, send, rpc, chainid, callback).
	private _buildLink(wallet: Wallet): string {
		const enc = encodeURIComponent;
		const parts: string[] = [
			`path=${enc(this._pkgPath)}`,
			`func=${enc(this._funcName)}`,
		];
		for (const [name, value] of this._readArgs()) {
			parts.push(`arg.${enc(name)}=${enc(value)}`);
		}
		const send = this._readSend();
		if (send) parts.push(`send=${enc(send)}`);

		const rpc = this._meta("gnoconnect:rpc");
		const chainid = this._meta("gnoconnect:chainid");
		if (rpc) parts.push(`rpc=${enc(rpc)}`);
		if (chainid) parts.push(`chainid=${enc(chainid)}`);
		parts.push(`callback=${enc(this._callbackURL())}`);

		return `${wallet.scheme}://tx?${parts.join("&")}`;
	}

	private _openWallet(wallet: Wallet): void {
		window.location.href = this._buildLink(wallet);
	}

	private _onSubmit(event: Event): void {
		// Fall through to the native submit whenever external-wallet routing
		// doesn't apply (desktop QR is a deferred follow-up).
		if (this._hasInPageProvider()) return;
		if (!this._isMobile()) return;

		const wallets = this._wallets();
		if (wallets.length === 0) return;

		event.preventDefault();
		this._openChooser(wallets);
	}

	// Populate and show the page-level chooser dialog. Always shown, even for
	// a single wallet: "Continue in browser" is the only way back to the
	// native submit when the wallet isn't installed (a failed custom-scheme
	// launch is silent).
	private _openChooser(wallets: Wallet[]): void {
		const dialog = this.getGlobalTarget("chooser") as HTMLDialogElement | null;
		const list = this.getGlobalTarget("chooser-list");
		if (!dialog || !list) {
			this._openWallet(wallets[0]); // no dialog — fail open
			return;
		}

		list.textContent = "";
		wallets.forEach((wallet) => {
			const li = document.createElement("li");
			const btn = document.createElement("button");
			btn.type = "button";
			btn.className = "b-wallet-chooser__item";
			if (wallet.icon) {
				const img = document.createElement("img");
				img.src = wallet.icon;
				img.alt = "";
				img.className = "b-wallet-chooser__icon";
				btn.appendChild(img);
			}
			const label = document.createElement("span");
			label.textContent = wallet.name;
			btn.appendChild(label);
			btn.addEventListener("click", () => {
				dialog.close();
				this._openWallet(wallet);
			});
			li.appendChild(btn);
			list.appendChild(li);
		});

		// Assignment (not addEventListener) so reopening doesn't stack handlers
		// or submit a previously opened form.
		const browser = this.getGlobalTarget("chooser-browser");
		if (browser) {
			browser.onclick = () => {
				dialog.close();
				// Native submit; bypasses submit listeners, so no re-interception.
				(this.element as HTMLFormElement).submit();
			};
		}
		const cancel = this.getGlobalTarget("chooser-cancel");
		if (cancel) cancel.onclick = () => dialog.close();

		if (typeof dialog.showModal === "function") {
			dialog.showModal();
			this._centerInVisualViewport(dialog);
		} else {
			dialog.setAttribute("open", "");
		}
	}

	// showModal() centers the dialog in the layout viewport, but a zoomed
	// mobile page (e.g. iOS auto-zoom on sub-16px inputs) only shows part of
	// it, so the dialog can land half off-screen. Shift it to the center of
	// the visual viewport instead, tracking zoom/scroll while open.
	private _centerInVisualViewport(dialog: HTMLDialogElement): void {
		const vv = window.visualViewport;
		if (!vv) return;

		const center = () => {
			const root = document.documentElement;
			// Cap to the visible area: the inner's 90vw is layout-viewport based
			// and overflows the visible width once the page is zoomed.
			dialog.style.maxWidth = `${vv.width * 0.9}px`;
			dialog.style.maxHeight = `${vv.height * 0.9}px`;
			const dx = vv.offsetLeft + (vv.width - root.clientWidth) / 2;
			const dy = vv.offsetTop + (vv.height - root.clientHeight) / 2;
			dialog.style.transform = `translate(${dx}px, ${dy}px)`;
		};
		center();
		vv.addEventListener("resize", center);
		vv.addEventListener("scroll", center);
		dialog.addEventListener(
			"close",
			() => {
				vv.removeEventListener("resize", center);
				vv.removeEventListener("scroll", center);
				dialog.style.transform = "";
				dialog.style.maxWidth = "";
				dialog.style.maxHeight = "";
			},
			{ once: true },
		);
	}
}
